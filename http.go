package main

import (
  "os"
  "os/signal"
  "syscall"
  "net/http"
  "strings"
  "log"
  "github.com/stianeikeland/go-rpio"
  "strconv"
  "fmt"
  "time"
)

var (
  // Use mcu pin 10, corresponds to physical pin 19 on the pi
  HWpin = rpio.Pin(10)
  HWfor int64 = 0
  HWuntil int64 = 0
  HWstate bool = false   
)

func cleanup() {
  fmt.Println("cleanup")
}

func resp() string {
  var HWRemaining int64 = 0
  if HWuntil != 0 {
    HWRemaining = HWuntil - time.Now().Unix()
  } 
  return "{\n\"HWfor\":" + strconv.FormatInt(HWfor, 10) + "\n\"HWRemaining\":" + strconv.FormatInt(HWRemaining,10)   + "\n\"HWUntil\":" + strconv.FormatInt(HWuntil, 10) + "\n\"HWstate\":" + strconv.FormatBool(HWstate)  + "\n}" 
} 

func sayHello(w http.ResponseWriter, r *http.Request) {
  message := r.URL.Path
  message = strings.TrimPrefix(message, "/")
  message = "Hello " + message + "\n"
  message = message + resp()
  w.Write([]byte(message))
}

func hwstate(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte(resp()))
}

func hwon(w http.ResponseWriter, r *http.Request) {
  HWstate = true
  t, Err := strconv.ParseInt(r.FormValue("for"), 10, 64)
  if Err == nil {
    HWuntil = time.Now().Unix() + t 
    HWfor = t 
  } else {
    HWuntil = 0
    HWfor = 0
  } 
  w.Write([]byte(resp()))
}

func hwoff(w http.ResponseWriter, r *http.Request) {
  HWstate = false
  HWuntil = 0
  HWfor = 0
  w.Write([]byte(resp()))
}

func handler() {
  http.HandleFunc("/", sayHello)
  http.HandleFunc("/HW/on", hwon)
  http.HandleFunc("/HW/off", hwoff)
  http.HandleFunc("/HW/state", hwstate)
  fmt.Println("listening")
  for true {
    log.Fatal(http.ListenAndServe(":8080", nil))
  }
}

func scheduler() {
  for true {
    switch {
    case HWpin.Read() == 0 && HWuntil > time.Now().Unix(): 
      HWTurnOn()
      //HWstate = true 
    case HWpin.Read() == 1 && HWuntil < time.Now().Unix() && HWuntil > 0:
      HWTurnOff()
      HWstate = false
      HWuntil = 0
      HWfor = 0
    case HWpin.Read() == 0 && HWstate == true:
      HWTurnOn()
    case HWpin.Read() == 1 && HWstate == false: 
      HWTurnOff()
    }
    time.Sleep(50 * time.Millisecond) 
  }
}

func HWTurnOn() {
  fmt.Println(time.Now().String() + " On")
  HWpin.High()
}

func HWTurnOff() {
  fmt.Println(time.Now().String() + " Off") 
  HWpin.Low()
}


func main() {
  rpio.Open()
  defer rpio.Close() // Unmap gpio memory when done
  HWpin.Output()
  go handler()
  go scheduler()
  c := make(chan os.Signal)
  signal.Notify(c, os.Interrupt, syscall.SIGTERM)
  go func() {
      <-c
      cleanup()
      os.Exit(0)
  }()

  for true {
      time.Sleep(10 * time.Second) 
  }
}
