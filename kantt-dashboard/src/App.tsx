import { useEffect, useState } from 'react'
import "./App.css";

const MINUTE = 60 * 1000
const HOUR = 60 * MINUTE

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

interface Task {
    name: string;
    startDate: string;
    endDate: string;
    color: string;
}


const mapPodToTask =  ({pods, startTime, endTime}: {pods: Pod[], startTime: number, endTime: number}): Task[] => {

    const colors = [
        "#4CAF50",
        "#2196F3",
        "#FF9800",
        "#9C27B0",
        "#00BCD4",
        "#795548",
        "#FFC107",
        "#607D8B",
    ]


    let result: Task[] = []
    for (let pod of pods) {
        let name = pod.name.String
        let startDate = new Date(Math.max(new Date(pod.createtime.Time).getTime(), startTime)).toISOString()
        let endDate = pod.deletetime.Valid ? pod.deletetime.Time : new Date(endTime).toISOString()
        let color = colors[result.length % colors.length]
        result.push({name, startDate, endDate, color})
    }

     return result
}


function ChartContainer () {
    const [startTime, setStartTime] = useState(Date.now() - HOUR);
    const [endTime, setEndTime] = useState(Date.now());


    return (
        <div>
            <input type="datetime-local" defaultValue={new Date(startTime).toISOString().slice(0,-1)} className='start-time-input'/>
            <input type="datetime-local" defaultValue={new Date(endTime).toISOString().slice(0, -1)} className='end-time-input'/>
            <button onClick={() => {
                const startTimeInput = document.querySelector('.start-time-input') as HTMLInputElement
                const endTimeInput = document.querySelector('.end-time-input') as HTMLInputElement
                setStartTime(new Date(startTimeInput.value).getTime())
                setEndTime(new Date(endTimeInput.value).getTime())
            }}>Submit</button>
            <Chart startTime={startTime} endTime={endTime} />
        </div>
    )

}

function Chart ({startTime, endTime}: {startTime: number, endTime: number}) {
    const [tasks, setTasks] = useState<Task[]>([])

    useEffect(() => {
        async function fetchData() {
            const response = await fetch(
                `${SERVER_URL}/pods?startTime=${startTime}&endTime=${endTime}`,
                {mode: 'cors'}
            )
            const result = await response.json()
            setTasks(mapPodToTask({pods: result.pods, startTime, endTime}))
        }
        fetchData()
    }, [startTime, endTime])


    const minTime = new Date(Math.min(...tasks.map(task => new Date(task.startDate).getTime())));
    const maxTime = new Date(Math.max(...tasks.map(task => new Date(task.endDate).getTime())));
    const totalTime = maxTime.getTime() - minTime.getTime();
    const chartWidth = 0.95 * window.innerWidth; // Chart takes 95% of the screen width


    return (
        <div className="gantt-container">
            <div className="gantt-body" style={{ width: chartWidth }}>
                {tasks.map((task) => {
                    const taskStart = new Date(task.startDate).getTime();
                    const taskEnd = new Date(task.endDate).getTime();
                    const offset = ((taskStart - minTime.getTime()) / totalTime) * (chartWidth *0.8);
                    const width = ((taskEnd - taskStart) / totalTime) * (chartWidth * 0.8);
                    return (
                        <div className="gantt-row" key={task.name}>
                            <div className="gantt-task-name">{task.name}</div>
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
                })}
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
