#!/usr/bin/env bash

set -euo pipefail

OUTPUT_FILE="$(cd "$(dirname "${1:-govulncheck-results.md}")" && pwd)/$(basename "${1:-govulncheck-results.md}")"

echo "# Go Vulnerability Check Results" > "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "Generated on: $(date)" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

GO_MOD_DIRS=$(find . -name "go.mod" -exec dirname {} \; | sort)

if [ -z "$GO_MOD_DIRS" ]; then
    echo "No Go modules found." >> "$OUTPUT_FILE"
    exit 0
fi

for DIR in $GO_MOD_DIRS; do
    echo "Checking $DIR..."
    pushd "$DIR" > /dev/null
    
    MODULE_NAME=$(grep "^module " "go.mod" | awk '{print $2}')
    
    echo "" >> "$OUTPUT_FILE"
    echo "## $MODULE_NAME" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    echo "Directory: \`$DIR\`" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    
    TEMP_OUTPUT=$(mktemp)
    govulncheck ./... > "$TEMP_OUTPUT" 2>&1 || true
    
    # Extract the summary lines and format them
    if grep -q "vulnerabilities from the Go standard library" "$TEMP_OUTPUT"; then
        STDLIB_LINE=$(grep "vulnerabilities from the Go standard library" "$TEMP_OUTPUT")
        echo "ℹ️ **Standard Library**: $STDLIB_LINE" >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
    fi
    
    if grep -q "This scan found no other vulnerabilities" "$TEMP_OUTPUT"; then
        echo "✅ **Dependencies**: No vulnerabilities found in imported packages or required modules." >> "$OUTPUT_FILE"
    elif grep -q "This scan also found" "$TEMP_OUTPUT"; then
        # Extract the complete multi-line summary
        DEP_SUMMARY=$(grep -A 2 "This scan also found" "$TEMP_OUTPUT" | tr '\n' ' ' | sed 's/  */ /g')
        echo "⚠️ **Dependencies**: $DEP_SUMMARY" >> "$OUTPUT_FILE"
    fi
    
    rm -f "$TEMP_OUTPUT"
    popd > /dev/null
done