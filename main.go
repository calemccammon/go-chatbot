package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"go-chatbot/internal/chat"
	"go-chatbot/internal/config"
	"go-chatbot/internal/llm"
)

const (
	windowTitle  = "Go Chatbot"
	windowWidth  = 900
	windowHeight = 680
)

type uiState struct {
	service        *chat.Service
	requestTimeout time.Duration

	chatLog     *widget.Entry
	status      *widget.Label
	promptInput *widget.Entry
	sendBtn     *widget.Button
	clearBtn    *widget.Button
}

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	client, err := llm.NewClient(cfg.APIKey, cfg.BaseURL, cfg.Model)
	if err != nil {
		log.Fatalf("llm client error: %v", err)
	}

	chatService, err := chat.NewService(client, cfg.SystemPrompt)
	if err != nil {
		log.Fatalf("chat service error: %v", err)
	}

	ui := newUI(chatService, cfg)
	ui.window.ShowAndRun()
}

type uiApp struct {
	app    fyne.App
	window fyne.Window
	state  *uiState
}

func newUI(service *chat.Service, cfg config.Config) uiApp {
	myApp := app.NewWithID("com.example.go-chatbot")
	myWindow := myApp.NewWindow(windowTitle)
	myWindow.Resize(fyne.NewSize(windowWidth, windowHeight))

	state := newUIState(service, cfg)
	content := buildContent(state)
	state.bindHandlers()

	myWindow.SetContent(content)

	return uiApp{
		app:    myApp,
		window: myWindow,
		state:  state,
	}
}

func newUIState(service *chat.Service, cfg config.Config) *uiState {
	chatLog := widget.NewMultiLineEntry()
	chatLog.SetPlaceHolder("Conversation appears here...")
	chatLog.Wrapping = fyne.TextWrapWord
	chatLog.Disable()

	status := widget.NewLabel("Ready")
	promptInput := widget.NewMultiLineEntry()
	promptInput.SetPlaceHolder("Ask something...")
	promptInput.Wrapping = fyne.TextWrapWord

	sendBtn := widget.NewButtonWithIcon("Send", theme.MailSendIcon(), nil)
	clearBtn := widget.NewButtonWithIcon("Clear", theme.DeleteIcon(), nil)

	return &uiState{
		service:        service,
		requestTimeout: cfg.RequestTimeout,
		chatLog:        chatLog,
		status:         status,
		promptInput:    promptInput,
		sendBtn:        sendBtn,
		clearBtn:       clearBtn,
	}
}

func buildContent(state *uiState) fyne.CanvasObject {
	controls := container.NewBorder(nil, nil, state.clearBtn, state.sendBtn, state.promptInput)
	return container.NewBorder(nil, state.status, nil, nil, container.NewVSplit(state.chatLog, controls))
}

func (s *uiState) bindHandlers() {
	s.sendBtn.OnTapped = s.send
	s.clearBtn.OnTapped = s.clear
	s.promptInput.OnSubmitted = func(_ string) {
		s.send()
	}
}

func (s *uiState) clear() {
	s.service.Clear()
	s.chatLog.SetText("")
	s.status.SetText("Cleared")
}

func (s *uiState) send() {
	prompt := strings.TrimSpace(s.promptInput.Text)
	if prompt == "" {
		return
	}

	s.appendMessage("You", prompt)
	s.promptInput.SetText("")
	s.status.SetText("Thinking...")
	s.setBusy(true)
	defer s.setBusy(false)

	ctx, cancel := context.WithTimeout(context.Background(), s.requestTimeout)
	defer cancel()

	reply, err := s.service.Send(ctx, prompt)
	if err != nil {
		s.status.SetText("Request failed")
		s.appendMessage("System", "Error: "+err.Error())
		return
	}

	s.appendMessage("Assistant", reply)
	s.status.SetText("Ready")
}

func (s *uiState) appendMessage(role, msg string) {
	current := strings.TrimSpace(s.chatLog.Text)
	line := fmt.Sprintf("%s: %s", role, strings.TrimSpace(msg))
	if current == "" {
		s.chatLog.SetText(line)
		return
	}

	s.chatLog.SetText(current + "\n\n" + line)
}

func (s *uiState) setBusy(busy bool) {
	if busy {
		s.sendBtn.Disable()
		s.promptInput.Disable()
		return
	}

	s.sendBtn.Enable()
	s.promptInput.Enable()
}
