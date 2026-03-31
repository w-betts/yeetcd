#!/bin/bash
# Open the human-facing documentation in a browser

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOC_PATH="$SCRIPT_DIR/documentation/human/index.html"

if [ ! -f "$DOC_PATH" ]; then
    echo "Error: Documentation not found at $DOC_PATH"
    echo "Please run the documentation agent first to generate the documentation."
    exit 1
fi

# Detect the operating system and use the appropriate command
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    open "$DOC_PATH"
elif [[ "$OSTYPE" == "linux"* ]]; then
    # Linux
    if command -v xdg-open &> /dev/null; then
        xdg-open "$DOC_PATH"
    else
        echo "Error: xdg-open not found. Please install it or manually open: $DOC_PATH"
        exit 1
    fi
else
    echo "Error: Unsupported operating system: $OSTYPE"
    echo "Please manually open: $DOC_PATH"
    exit 1
fi
