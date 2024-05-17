ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR/docs" || { echo "Failed to navigate to repository root"; exit 1; }
rm demo.svg 2> /dev/null
svg-term --in demo.cast --from 22000 --out demo.svg --term terminal --no-optimize