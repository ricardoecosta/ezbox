# ezbox
Elderly friendly media center based on the Raspberry Pi platform.

GPIO ports are used via physical switches to change channels and change programs back and forth.

Using file descriptors to GPIO ports state changes which are streamed to a websocket via go channels.

Basic ELM client listening on the websocket stream and reacting to events triggered by the GPIO controls. 

Using `omxplayer` for playing media files and leverage full media playback capabilities from the hardware.

# todo
* tests
* uniform error handling
* speed up rotary encoder via mem access
* auto play start/stop
* sanitize video file names
* rebuild media collection index when directory changes
* install hardware clock module