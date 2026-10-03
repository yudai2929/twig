# Build ./bin/twig before recording this tape.
Output docs/demo.gif
Output docs/demo.mp4

Set Shell "zsh"
Set FontSize 20
Set Width 1200
Set Height 470
Set Framerate 20
Set TypingSpeed 170ms
Set Theme "Dracula"

Hide
Type "zsh -f"
Enter
Sleep 300ms
Type "export TWIG_BIN=./bin/twig; source ./zsh/twig.zsh; PROMPT='twig> '; RPROMPT=''"
Enter
Sleep 1s
Ctrl+L
Sleep 300ms
Show

Type "gi"
Sleep 1s
Enter
Sleep 1s
Type "c"
Sleep 1s
Down@350ms 5
Sleep 500ms
Enter
Sleep 2s
