package main

func exampleObserverPattern() {
	println("exampleObserverPattern")
	// 创建观察者WebGUI、AlertSystem
	obs := []Observer{&WebGUI{}, &AlertSystem{}}

	// 创建一个天气服务
	ws := NewWeatherService()
	ws.Register(obs...)
	ws.ChangeWeather("Sunny") // 天气变化
	ws.Notify()               // 通知所有观察者

	ws.ChangeWeather("Rain")
	ws.Remove(obs[0])
	ws.Notify()
}

// 主题（接口）

type WeatherServiceSubject interface {
	Register(observer ...Observer) // 注册观察者
	Remove(observer Observer)      // 注销观察者
	Notify()                       // 通知所有注册的观察者
	ChangeWeather(string2 string)  // 天气变化
}

// 观察者（接口）

type Observer interface {
	Update(weather string) // 观察者需要实现这个方法来接收天气信息的更新
}

// 具体主题：天气服务

type WeatherService struct {
	observers map[Observer]struct{} // 观察者集合
	weather   string                // 天气情况
}

var _ WeatherServiceSubject = (*WeatherService)(nil)

func (w *WeatherService) Register(observer ...Observer) {
	// 为了简便，这里没有处理并发
	for _, ob := range observer {
		w.observers[ob] = struct{}{}
	}
}

func (w *WeatherService) Remove(observer Observer) {
	delete(w.observers, observer)
}

func (w *WeatherService) Notify() {
	for k := range w.observers {
		k.Update(w.weather)
	}
}

func (w *WeatherService) ChangeWeather(weather string) {
	w.weather = weather
}

func NewWeatherService() WeatherServiceSubject {
	return &WeatherService{
		observers: make(map[Observer]struct{}),
	}
}

// WebGUI 作为具体观察者，负责更新网页显示
type WebGUI struct {
}

var _ Observer = (*WebGUI)(nil)

func (wg *WebGUI) Update(weather string) {
	//fmt.Printf("WebGUI: Updating the display to show weather: %s\n", weather)
}

// AlertSystem 作为具体观察者，负责发送预警短信
type AlertSystem struct {
}

var _ Observer = (*AlertSystem)(nil)

func (as *AlertSystem) Update(weather string) {
	//if weather == "Rain" {
	//	fmt.Printf("AlertSystem: Sending rain alert!\n")
	//}
}
