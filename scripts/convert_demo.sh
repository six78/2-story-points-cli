ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR/docs" || { echo "Failed to navigate to repository root"; exit 1; }
rm demo.svg 2> /dev/null
rm demo.gif 2> /dev/null
svg-term --in demo.cast --from 22000 --out demo.svg --term terminal --no-optimize

#asciinema-edit cut --start=1.921224 --end=22.56033 --out demo-cut.cast demo.cast
# TODO: replace first line from original file, or face `Error: ParseJson(Error("missing field `fg`", line: 1, column: 70))`

agg demo.cast demo.gif --font-family="JetBrains Mono" --font-size=14 --idle-time-limit=1 --theme=dracula

# TODO: cut first 22 seconds of GIF with https://ezgif.com/cut/

