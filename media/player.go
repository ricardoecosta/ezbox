package media

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

const PlayerWrapper = "omxplayer"
const PlayerBinary = "omxplayer.bin"

// todo: kill any running omxplayer process
// todo: channel for process status
// todo: graceful shutdown

var (
	stdin   io.WriteCloser
	cmd     *exec.Cmd
	running = false
	mutex   sync.Mutex
)

func InitPlayer(omxplayerPath string) {
	if omxplayerPath != "" {
		_, err := os.Stat(omxplayerPath)
		if err == nil {
			log.Println("using omxplayer located at: " + omxplayerPath)
		} else {
			log.Panicf("omxplayer not found at configured location: " + omxplayerPath)
		}
	}

	log.Println("looking up omxplayer in $PATH")
	playerPath := lookupPlayerInPath()

	if playerPath != "" {
		log.Println("found omxplayer at: " + playerPath)
	} else {
		log.Println("omxplayer not found in $PATH, media player capability disabled")
	}
}

func lookupPlayerInPath() string {
	path, err := exec.LookPath(PlayerWrapper)
	if err != nil {
		return ""
	} else {
		return path
	}
}

func PlayFile(file string) error {
	return startPlayerWithArgs(file, "-o", "alsa", "--aspect-mode", "fill", "-b", "-w")
}

func FileIsSupported(file string) bool {
	output, _ := exec.Command(PlayerWrapper, fmt.Sprintf("%s", file), "--info").CombinedOutput()
	return strings.Contains(string(output), file)
}

func startPlayerWithArgs(file string, args ...string) error {
	mutex.Lock()
	defer mutex.Unlock()

	cmd = exec.Command(PlayerWrapper, append(args, file)...)

	var err error
	stdin, err = cmd.StdinPipe()
	if err != nil {
		return errors.WithMessage(err, "unable to get hold of omxplayer stdin pipe")
	}

	err = cmd.Start()
	if err != nil {
		return errors.WithMessage(err, "unable to start omxplayer")
	}
	running = true
	return nil
}

func Stop() {
	mutex.Lock()
	defer mutex.Unlock()

	if running {
		_, err := fmt.Fprint(stdin, "q")
		if err == nil {
			stdin.Close()
			cmd.Wait()
		} else {
			exec.Command("killall " + PlayerBinary).Run()
		}
		running = false
	}
}
