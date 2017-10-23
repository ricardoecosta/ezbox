package main

import (
	"github.com/ricardoecosta/ezbox/gpio"
)

type Metadata struct {
	Type string `json:"type"`
}

// Outbound message types

type PinUpdated struct {
	Metadata
	Pin gpio.Pin `json:"pin"`
}

type ChannelChanged struct {
	Metadata
	Channel    string `json:"channel"`
	Title      string `json:"title"`
	AudioTitle string `json:"audio_title"`
}

type ProgramUpdated struct {
	Metadata
	Program string `json:"program"`
}

type ProgramPaused struct {
	Metadata
	Pause bool `json:"pause"`
}

// Inbound message types, simulated gpio mode

type SimulatePinUpdate struct {
	Metadata
	Pin gpio.Pin `json:"pin"`
}

// Message factories

func NewPinUpdatedMessage(pin gpio.Pin) PinUpdated {
	return PinUpdated{
		Metadata: Metadata{Type: "PinUpdated"},
		Pin:      pin,
	}
}

func NewChannelChangedMessage(channel string, title string, audioTitle string) ChannelChanged {
	return ChannelChanged{
		Metadata:   Metadata{Type: "ChannelChanged"},
		Channel:    channel,
		Title:      title,
		AudioTitle: audioTitle,
	}
}

func NewProgramUpdatedMessage(program string) ProgramUpdated {
	return ProgramUpdated{
		Metadata: Metadata{Type: "ProgramUpdated"},
		Program:  program,
	}
}

func NewProgramPausedMessage(pause bool) ProgramPaused {
	return ProgramPaused{
		Metadata: Metadata{Type: "ProgramPaused"},
		Pause:    pause,
	}
}

func NewSimulatePinMessage(pin gpio.Pin) SimulatePinUpdate {
	return SimulatePinUpdate{
		Metadata: Metadata{Type: "SimulatePin"},
		Pin:      pin,
	}
}
