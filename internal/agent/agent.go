package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	RequestPort          string `env:"SERVER_PORT"`
	TimeAdditionMs       int    `env:"TIME_ADDITION_MS"`
	TimeSubtractionMs    int    `env:"TIME_SUBTRACTION_MS"`
	TimeMultiplicationMs int    `env:"TIME_MULTIPLICATIONS_MS"`
	TimeDivisionMs       int    `env:"TIME_DIVISIONS_MS"`
	ComputingPower       int    `env:"COMPUTING_POWER"`
}

func GetCfgFromEnv() *Config {
	var cfg Config

	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		log.Printf("Error loading config, loaded default values: %v", err)
		return &Config{
			TimeAdditionMs:       2000,
			TimeSubtractionMs:    2000,
			TimeMultiplicationMs: 5000,
			TimeDivisionMs:       5000,
			ComputingPower:       4,
		}
	}

	log.Printf("Loaded config, %v:", cfg)
	return &cfg
}

type Agent struct {
	cfg           *Config
	taskChan      chan *GetTaskResponse
	readyTaskChan chan *SendResultResponce
	wg            sync.WaitGroup
}

func New() *Agent {
	return &Agent{
		cfg:           GetCfgFromEnv(),
		taskChan:      make(chan *GetTaskResponse, 100),    // для получения задач
		readyTaskChan: make(chan *SendResultResponce, 100), // для результатов
	}
}

type GetTaskResponse struct {
	ID        string `json:"id"`
	Arg1      string `json:"arg1"`
	Arg2      string `json:"arg2"`
	Operation string `json:"operation"`
}

// Получение задачи
func (a *Agent) GetTask() *GetTaskResponse {
	resp, err := http.Get("http://localhost:" + a.cfg.RequestPort + "/api/v1/task")
	if err != nil {
		fmt.Println("Failed to get task:", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(time.Now(), "Error with status", resp.StatusCode)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed ro read response:", err)
		return nil
	}

	var task GetTaskResponse
	err = json.Unmarshal(body, &task)
	if err != nil {
		fmt.Println("Failed to Unmarshal JSON:", err)
		return nil
	}

	return &task
}

type SendResultResponce struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
	Error  string  `json:"error"`
}

// Отправление задачи
func (a *Agent) SendResult(readyTask *SendResultResponce) {
	jsonTask, err := json.Marshal(readyTask)
	if err != nil {
		log.Printf("Failed to marshal task: %v", err)
		return
	}

	req, err := http.NewRequest("POST", "http://localhost:"+a.cfg.RequestPort+"/api/v1/task", bytes.NewBuffer(jsonTask))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request error, %v:", err)
		return
	}
	defer resp.Body.Close()
}

func (a *Agent) worker() {
	defer a.wg.Done()
	for task := range a.taskChan {
		a.readyTaskChan <- a.processTask(task)

		readyTask := <-a.readyTaskChan
		log.Println("Get ready task:", readyTask)

		a.SendResult(readyTask)
	}
}

func (a *Agent) processTask(task *GetTaskResponse) *SendResultResponce {
	log.Printf("Обработка задачи: ID=%s, Arg1=%s, Arg2=%s, Operation=%s\n", task.ID, task.Arg1, task.Arg2, task.Operation)

	// Нет смысла обрабатывать err потому что такого рода ошибки сюда не дойдут
	arg1, err := strconv.ParseFloat(task.Arg1, 64)
	if err != nil {
		return &SendResultResponce{ID: task.ID, Error: ErrInvalidOperator.Error()}
	}
	arg2, err := strconv.ParseFloat(task.Arg2, 64)
	if err != nil && task.Operation != "~" {
		return &SendResultResponce{ID: task.ID, Error: ErrInvalidOperator.Error()}
	}

	switch task.Operation {
	case "+":
		time.Sleep(time.Duration(a.cfg.TimeAdditionMs) * time.Millisecond)
		return &SendResultResponce{ID: task.ID, Result: arg1 + arg2}
	case "-":
		time.Sleep(time.Duration(a.cfg.TimeSubtractionMs) * time.Millisecond)
		return &SendResultResponce{ID: task.ID, Result: arg1 - arg2}
	case "*":
		time.Sleep(time.Duration(a.cfg.TimeMultiplicationMs) * time.Millisecond)
		return &SendResultResponce{ID: task.ID, Result: arg1 * arg2}
	case "/":
		if arg2 == 0 {
			return &SendResultResponce{ID: task.ID, Error: ErrDevisionByZero.Error()}
		}
		time.Sleep(time.Duration(a.cfg.TimeDivisionMs) * time.Millisecond)
		return &SendResultResponce{ID: task.ID, Result: arg1 / arg2}
	case "~":
		time.Sleep(time.Duration(a.cfg.TimeSubtractionMs) * time.Millisecond)
		return &SendResultResponce{ID: task.ID, Result: -arg1}
	default:
		log.Println("Invalid operation")
		return &SendResultResponce{ID: task.ID, Error: ErrInvalidOperation.Error()}
	}
}

func (a *Agent) RunAgent() error {
	// Запуск воркеров
	for i := 0; i < a.cfg.ComputingPower; i++ {
		a.wg.Add(1)
		go a.worker()
	}

	// Бесконечный цикл для запроса задач
	go func() {
		for {
			task := a.GetTask()
			if task != nil {
				a.taskChan <- task
			}
			time.Sleep(1 * time.Second) // Задержка между запросами задач
		}
	}()

	a.wg.Wait()
	return nil
}
