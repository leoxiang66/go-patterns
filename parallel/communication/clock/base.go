package clock

type ClockInterface interface {
	Tick()
	Time() int
}
