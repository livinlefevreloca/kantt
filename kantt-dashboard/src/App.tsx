import { useEffect, useState, useLayoutEffect, useRef } from 'react'

import dayjs from 'dayjs'
import minMax from 'dayjs/plugin/minMax'
dayjs.extend(minMax)

import "./App.css";

const MINUTE = 60
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR
const WEEK = 7 * DAY
const MONTH = 30 * DAY
const YEAR = 365 * DAY
const TIME_BLOCKS = 7

const SERVER_URL = 'http://localhost:8080'

interface StringField {
    String: string
    Valid: boolean
}

interface TimeField {
    Time: string
    Valid: boolean
}

interface Pod {
    name: StringField
    namespace: StringField
    createtime: TimeField
    deletetime: TimeField
    ownername: StringField
    ownerkind: StringField
    nodename: StringField
    nodeip: StringField
}

interface Node {
    name: StringField
}

interface Namespace {
    namespace: StringField
}

interface Task {
    name: string;
    namespace: string;
    nodename: string;
    startDate: dayjs.Dayjs;
    endDate: dayjs.Dayjs;
    color: string | null;
}


const mapPodToTask =  ({pods, startTime, endTime}: {pods: Pod[], startTime: dayjs.Dayjs, endTime: dayjs.Dayjs}): {[key: string]: Task[]} => {
    const colors = [
        "#3944BC",
        "#757C88",
        "#63C5DA",
        "#151E3D",
        "#0492C2",
        "#1338BE",
        "#52B2BF",
        "#241571",
    ]
    let groups: {[key: string]: Task[]} = {}
    for (let pod of pods) {
        let namespace = pod.namespace.String
        let task = {
            name: pod.name.String,
            namespace: pod.namespace.String,
            nodename: pod.nodename.String,
            startDate: dayjs.max(dayjs(pod.createtime.Time), startTime),
            endDate: pod.deletetime.Valid ? dayjs(pod.deletetime.Time): endTime,
            color: null
        }

        if (!groups[namespace]) {
            groups[namespace] = []
        }
        groups[namespace].push(task)
    }
    for (let group in groups) {
        groups[group] = groups[group].sort((a, b) => a.name.localeCompare(b.name))
        let colorsIndex = 0
        for (let task of groups[group]) {
            task.color = colors[colorsIndex % colors.length]
            colorsIndex++
        }
    }

    return groups
}

const calcMinTime = (tasks: {[key: string]: Task[]}): dayjs.Dayjs => {
    let values = []
    for (let group in tasks) {
        if (tasks[group].length === 0) {
            continue
        }
        values.push(tasks[group].map((task) => task.startDate).reduce((acc, startDate) => dayjs.min(acc, startDate)))
    }
    return dayjs.min(values)

}

const calcMaxTime = (tasks: {[key: string]: Task[]}): dayjs.Dayjs => {
    let values = []
    for (let group in tasks) {
        values.push(tasks[group].map((task) => task.endDate).reduce((acc, endDate) => dayjs.max(acc, endDate)))
    }
    return dayjs.max(values)
}


const resolveLabelFromTimeBlock = (timeBlock: number): string => {
    if (timeBlock < MINUTE) {
        return 'MMM-DD, HH:mm:ss';
    } else if (timeBlock < HOUR) {
        return 'MMM-DD, HH:mm:ss';
    } else if (timeBlock < DAY) {
        return 'MMM-DD, HH:mm';
    } else if (timeBlock < WEEK) {
        return 'MMM DD HH:mm';
    } else if (timeBlock < MONTH) {
        return 'MMM DD, YYYY';
    } else if (timeBlock < YEAR) {
        return 'MMM, YYYY';
    } else {
        return 'YYYY';
    }
}


const renderTimeAxis = (minTime: dayjs.Dayjs, maxTime: dayjs.Dayjs, chartWidth: number): JSX.Element => {
    const rowWidth = (chartWidth * 0.8) -10;
    const totalTime = maxTime.unix() - minTime.unix();
    const timeBlockSize = totalTime / TIME_BLOCKS;
    const formatStr = resolveLabelFromTimeBlock(timeBlockSize);
    return (
        <div className="gantt-time-axis" style={{ width: chartWidth }}>
            <div className="gantt-time-axis-padding" style={{width: chartWidth * 0.2}}></div>
            <div className="gantt-time-axis-dateline" style={{width: rowWidth}}>
                {Array.from({ length: TIME_BLOCKS }, (_, i) => {
                    const time = minTime.add(timeBlockSize * i, 'second');
                    return (
                        <div className="gantt-time-axis-dateline-item" key={time.unix()}>
                            {time.format(formatStr)}
                        </div>
                    );
                })}
            </div>
        </div>
    );
}

const renderTasks = (tasks: Task[], totalTime: number, minTime: dayjs.Dayjs, chartWidth: number): JSX.Element[] => {
    return tasks.map((task) => {
        const rowWidth = (chartWidth * 0.8) - 10;
        const taskStart = task.startDate.unix();
        const taskEnd = task.endDate.unix();
        const offset = ((taskStart - minTime.unix()) / totalTime) * rowWidth;
        const width = ((taskEnd - taskStart) / totalTime) * rowWidth;
        return (
            <div className="gantt-row" key={`${task.namespace}-${task.name}-${task.nodename}`}>
                <div className="gantt-task-name">{`${task.namespace}/${task.name}`}</div>
                <div className="gantt-task-wrapper">
                    <div
                        className="gantt-task"
                        style={{
                            backgroundColor: task.color,
                            marginLeft: `${offset}px`,
                            width: `${width}px`,
                        }}
                    ></div>
                </div>
            </div>
        );
        })
}

const renderGroups = (groups: {[key: string]: Task[]}, totalTime: number, minTime: dayjs.Dayjs, chartWidth: number): JSX.Element[] => {

    return Object.keys(groups).map((group: string) => {
        return (
            <div className="gantt-group" key={group}>
                <div className="gantt-group-name">{group}</div>
                <div className="gantt-tasks">
                    {renderTasks(groups[group], totalTime, minTime, chartWidth)}
                </div>
            </div>
        );
    })

}

interface windowSize {
    width: number,
    height: number
}

function ChartContainer () {


    // The currently selected namespace and node used for
    // filtering the choices for the dropdowns
    const [selectedNamespace, setSelectedNamespace] = useState('')
    const [selectedNodeName, setSelectedNodeName] = useState('')
    const [selectedStartTime, setSelectedStartTime] = useState<dayjs.Dayjs | null>(null)
    const [selectedEndTime, setSelectedEndTime] = useState<dayjs.Dayjs | null>(null)

    useEffect(() => {
        async function fetchData() {
            let startTime = selectedStartTime ? selectedStartTime.unix() : null
            let endTime = selectedEndTime ? selectedEndTime.unix() : null

            let requestUrl = `${SERVER_URL}/namespaces?startTime=${startTime}&endTime=${endTime}`
            if (selectedNodeName) {
                requestUrl += `&node=${selectedNodeName}`
            }
            const response = await fetch(
                requestUrl,
                {mode: 'cors'}
            )
            const result = await response.json()
            let namespaces = result.namespaces.filter((namespace: Namespace) => namespace.namespace.Valid).map((namespace: Namespace) => namespace.namespace.String)
            namespaces.unshift('all')
            setNamespaces(namespaces)
        }
        if (!selectedStartTime || !selectedEndTime) {
            return
        }
        fetchData()
    }, [selectedNodeName, selectedStartTime, selectedEndTime])

    useEffect(() => {
        async function fetchData() {

            let startTime = selectedStartTime ? selectedStartTime.unix() : null
            let endTime = selectedEndTime ? selectedEndTime.unix() : null

            let requestUrl = `${SERVER_URL}/nodes?startTime=${startTime}&endTime=${endTime}`

            if (selectedNamespace) {
                requestUrl += `&namespace=${selectedNamespace}`
            }

            const response = await fetch(
                requestUrl,
                {mode: 'cors'}
            )
            const result = await response.json()
            let nodes = result.nodes.filter((node: Node) => node.name.Valid).map((node: Node) => node.name.String)
            nodes.unshift('all')
            setNodes(nodes)
        }
        if (!selectedStartTime || !selectedEndTime) {
            return
        }
        fetchData()
    }, [selectedNamespace, selectedStartTime, selectedEndTime])

    // Choices for namespace and node
    const [nodes, setNodes] = useState<string[]>([])
    const [namespaces, setNamespaces] = useState<string[]>([])


    // Selected Values to pass tot the chart on submit
    const [nodeName, setNodeName] = useState('')
    const [namespace, setNamespace] = useState('')
    const [startTime, setStartTime] = useState<dayjs.Dayjs | null>(null);
    const [endTime, setEndTime] = useState<dayjs.Dayjs | null>(null);


    return (
        <div>
            <div className='inputs-container'>
                <input
                    type="datetime-local"
                    onChange={(e) => setSelectedStartTime(dayjs(e.target.value))}
                    className='start-time-input'/>
                <input
                    type="datetime-local"
                    onChange={(e) => setSelectedEndTime(dayjs(e.target.value))}
                    className='end-time-input'
                />
                <select onChange={(e) => setSelectedNamespace(e.target.value)} className='namespace-input'>
                    {namespaces.length > 0 ? namespaces.map((namespace) => <option key={namespace} value={namespace}>{namespace}</option>) : <option value='all'>all</option>}
                </select>
                <select onChange={(e) => setSelectedNodeName(e.target.value)} className='node-input'>
                    {nodes.length > 0 ? nodes.map((node) => <option key={node} value={node}>{node}</option>) : <option value='all'>all</option>}
                </select>
                <button onClick={() => {
                    if (!selectedStartTime || !selectedEndTime) {
                        alert('Please select a start and end time')
                        return
                    }
                    setStartTime(selectedStartTime)
                    setEndTime(selectedEndTime)
                    setNamespace(selectedNamespace)
                    setNodeName(selectedNodeName)
                }}
                className='submit-button'
                >Submit</button>

            </div>
            <Chart startTime={startTime} endTime={endTime} namespace={namespace} node={nodeName}/>
        </div>
    )

}

interface ChartProps {
    startTime: dayjs.Dayjs | null,
    endTime: dayjs.Dayjs | null,
    namespace: string,
    node: string,
}

function Chart ({startTime, endTime, namespace, node}: ChartProps) {
    const [size, setSize] = useState({width: window.innerWidth, height: window.innerHeight} as windowSize);
    useLayoutEffect(() => {
    function updateSize() {
    setSize({width: window.innerWidth, height: window.innerHeight});
    }
    window.addEventListener('resize', updateSize);
    updateSize();
    return () => window.removeEventListener('resize', updateSize);
    }, []);


    const [tasks, setTasks] = useState<{[key: string]: Task[]}>({})

    useEffect(() => {
        async function fetchData() {
            if (!startTime || !endTime) {
                return
            }
            let requestUrl = `${SERVER_URL}/pods?startTime=${startTime.unix()}&endTime=${endTime.unix()}`
            if (namespace) {
                requestUrl += `&namespace=${namespace}`
            }
            if (node) {
                requestUrl += `&node=${node}`
            }

            const response = await fetch(
                requestUrl,
                {mode: 'cors'}
            )
            const result = await response.json()
            setTasks(mapPodToTask({pods: result.pods, startTime, endTime}))
        }
        fetchData()
    }, [startTime, endTime, namespace, node])

    if (!startTime || !endTime || Object.keys(tasks).length === 0) {
        return <div className='gantt-container' style={{color: 'black'}}><h2>Such Empty =(</h2></div>
    }

    const chartWidth = 0.95 * size.width + 10;
    const minTime = calcMinTime(tasks);
    const maxTime = calcMaxTime(tasks);
    const totalTime = maxTime.unix() - minTime.unix();


    return (
        <div className="gantt-container">
            {renderTimeAxis(minTime, maxTime, chartWidth)}
            <div className="gantt-body" style={{ width: chartWidth}}>
                {renderGroups(tasks, totalTime, minTime, chartWidth)}
            </div>
        </div>
    );

}


function App() {
  return (
    <>
        <ChartContainer />
    </>
  )
}


export default App
