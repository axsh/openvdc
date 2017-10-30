package scheduler

type SchedulingMethod int

const (
        None SchedulingMethod = iota // 0
        RoundRobin
)

func (m SchedulingMethod) String() string {
        switch m {
        case RoundRobin:
                return "round-robin"
        default:
                return "none"
        }
}
