# Mouser

> Let your mouse do your work for you. Automate actions via simple mouse gestures.

Mouser currently supports macOS only. See [Development](#development) if you'd like to help add support for other OSes.

## Contents

- [Features](#features)
- [Installation](#installation)
  - [macos](#macos)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)
- [Development](#development)
- [Credits](#credits)


## Features

### Gestures
- **Presses:** short press, long press, key down, key up
- **Swipes:** swipe up/down/left/right
- **Combined Gestures:** double/triple tap (two/three consecutive short presses), swipe pattern (e.g. swipe left > swipe right > swipe left)

### Actions
- **Basic Actions:** control volume, media playback, trigger shortcuts, type text, run commands, etc.
- **Toggle Actions:** repeat actions until stopped (e.g. repeat action while key is pressed)
- **App-Specific Actions:** trigger a different action based on the current app


## Installation

Below you'll find the recommended ways to install Mouser.

Alternatively, you can download Mouser from the [Releases](https://github.com/birdkid/Mouser/releases) page, or refer to [Development](#development) to build it yourself.

### macOS
Via [Homebrew](https://brew.sh/):
```sh
# Install:
brew install birdkid/tap/mouser
# Update:
brew upgrade birdkid/tap/mouser
# Auto-start:
brew services start mouser
```


## Configuration

Mouser uses a YAML file for its configuration.

Default config file paths:
- **macOS:** `~/.config/mouser/config.yml`

Alternatively, you can specify a custom path:
```sh
mouser --config path/to/config/file.yml
```

The configuration file consists of these sections:
- `mappings`: Lists optional aliases for keys and buttons.
- `gestures`: Maps keys/buttons/aliases and gestures to actions.
- `actions`: Contains custom actions.
- `settings`: Controls additional settings.

### Example Configurations

<details open>
<summary title="View Simple Configuration"><strong>Simple Configuration</strong></summary>

```yaml
gestures:
  mouse4:
    swipe_left: media:prev
    swipe_right: media:next
    tap: media:toggle
  f19:
    tap:
      action: io:tap
      args: [cmd, w]
    hold:
      action: io:tap
      args: [cmd, q]
```
</details>

<details>
<summary title="View Advanced Configuration"><strong>Advanced Configuration</strong></summary>

```yaml
mappings:
  GOTO_PREV: f13
  GOTO_NEXT: f15
  VOL_DOWN: f11
  VOL_UP: f12
  MEDIA: mouse4
  CLOSE: mouse5

gestures:
  GOTO_PREV:
    key_down: mac:prev-tab:toggle:on
    key_up: mac:prev-tab:toggle:off
  GOTO_NEXT:
    key_down: mac:next-tab:toggle:on
    key_up: mac:next-tab:toggle:off
  VOL_DOWN:
    key_down: vol:down:toggle:on
    key_up: vol:down:toggle:off
  VOL_UP:
    key_down: vol:up:toggle:on
    key_up: vol:up:toggle:off
  MEDIA:
    swipe_left: media:prev
    swipe_right: media:next
    swipe_up: mac:open-media-player
    tap.tap: media:toggle
    hold: media:toggle
  CLOSE:
    tap: mac:smart-close-window
    hold: mac:quit-app

actions:

  mac:prev-tab:
    action: io:tap
    args: [ctrl, shift, tab]
  mac:next-tab:
    action: io:tap
    args: [ctrl, tab]

  vol:down:toggle:
    type: toggle
    action: vol:down
    init-delay: 100
    repeat-delay: 100
  vol:up:toggle:
    type: toggle
    action: vol:up
    init-delay: 100
    repeat-delay: 100

  mac:close-window:
    action: io:tap
    args: [cmd, w]
  mac:quit-app:
    action: io:tap
    args: [cmd, q]

  mac:open-media-player:
    action: os:open
    args: [/Applications/Spotify.app]

  mac:smart-close-window:
    type: app-branch
    branches:
      /Applications/MyCriticalApp.app: null
    fallback: mac:close-window

settings:
  toggles:
    init-delay: 250
    repeat-delay: 200
```
</details>

### Configuration Details

#### Gestures
<details>
<summary title="View Available Gestures">Available Gestures</summary>

- `key_down`
- `key_up`
- `tap`
- `hold`
- `swipe_up`
- `swipe_down`
- `swipe_left`
- `swipe_right`
</details>

<details>
<summary title="View Gestures Matching Pattern Examples">Gestures Matching Pattern Examples</summary>

- `swipe_up`: every `swipe_up` event.
- `tap.tap`: double-taps.
- `swipe_left.swipe_down.swipe_right.swipe_up`: when swiping ← ↓ → ↑.
</details>

#### Actions
<details>
<summary title="View Available Actions">Available Actions</summary>

- `vol:down`: decreases the audio volume level
- `vol:up`: increases the audio volume level
- `vol:mute`: toggles between muting and unmuting audio
- `media:toggle`: toggles between playing and pausing the current media
- `media:prev`: rewindes the current or jumps back to the previous media record
- `media:next`: forwards to the next media record
- `os:close-window`: closes the current window
- `misc:none`: does nothing

- `io:tap`: triggers a short key press & release; arguments:
	- _modifiers…_: optional modifiers to hold during the key tap, e.g.
	  `shift`, `cmd`, etc.
	- _key_: the name of the key to tap, e.g. `f1`, `a`, `enter` etc

- `io:type`: writes out the given text; arguments:
	- _text_: the text to type out

- `io:scroll`: triggers a scroll event; arguments:
	- _x_: the distance in pixels to scroll horizontally (left to right)
	- _y_: the distance in pixels to scroll vertically (top to bottom)

- `os:open`: opens a file or application; arguments:
	- _file_: the path to the file or application to open
	- _openArgs…_: list of extra arguments to pass to the open command

- `os:cmd`: runs a custom command; arguments:
	- _cmd_: the command name or path
	- _cmdArgs…_: list of extra arguments to pass to the command

- `misc:sleep`: pauses action execution for a given time; arguments:
	- _duration_: the duration of the pause in milliseconds > 0
</details>

<details>
<summary title="View Action Types">Action Types</summary>

```yaml
actions:

  # Simple actions.
  my-action: some-action
  my-action-with-args:
    action: some-action
    args: [12, 34]

  # Toggle actions.
  my-toggle-with-delay:
    type: toggle
    action: some-toggled-action
    init-delay: 500
  my-fast-toggle:
    type: toggle
    action:
      action: some-toggled-action
      args: [56, 78]
    init-delay: 50
    repeat-delay: 0

  # App-specific actions.
  my-app-actions:
    type: app-branch
    branches:
      /Applications/MyApp1.app: some-app1-action
      /Applications/MyApp2.app:
        action: some-app2-action
        args: [foo, bar]
    fallback: some-fallback-action
```
</details>

#### Settings
<details>
<summary title="View All Settings">All Settings</summary>

```yaml
# All times are in milliseconds.
# All distances are in pixels.
settings:

  # Enable verbose logging.
  debug: false

  gestures:
    # Min time before a subsequent gesture starts a new gesture combo.
    ttl: 500
    # Max time until a press-and-release is considered a short press ("tap")
    # instead of a long press ("hold").
    short-press-ttl: 500
    # Max number of gestures in a given gesture combo.
    cap: 8

  swipes:
    # Min distance until mouse movement is considered a swipe.
    min-dist: 30
    # Max time repetitive identical swipe directions are surpressed.
    throttle: 250
    # Tick rate determining how often the current mouse position is checked for
    # a potential swipe gesture.
    poll-rate: 100

  toggles:
    # Default delay between the first action trigger and subsequent repeats.
    init-delay: 200
    # Default delay between subsequent repeats.
    repeat-delay: 100
```
</details>


## Troubleshooting

### macOS
<details>
<summary title="View Error: monitor initialization failed"><strong>Error:</strong> <code>monitor initialization failed</code></summary>

Ensure you have granted the app the necessary permissions:
1. Go to _System Perferences > Security & Privacy > Privacy > Accessibility_.
1. Enable _mouser_.
1. Restart Mouser.
</details>


## Development

### Dev Requirements

- **All:** `Golang`, `GCC`
- **macOS:** `Xcode Command Line Tools`

### Dev Run/Test/Build
#### Run
```sh
make run
```
#### Test
```sh
make mock
make test
```
#### Build
```sh
make build
```


## Credits

- [RobotGo](https://github.com/go-vgo/robotgo)—used extensively in this project
- [MASShortcut](https://github.com/shpakovski/MASShortcut)—for references of hotkeys on macOS
