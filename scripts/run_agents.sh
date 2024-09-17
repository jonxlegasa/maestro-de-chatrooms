#!/bin/bash

# Default values
USERNAME_PREFIX="OPENAIGPT"
FILES_DIR=""
LISTEN_ADDR=""
CONNECT_ADDR="127.0.0.1:4000"

# Function to display usage
usage() {
    echo "Usage: $0 [-p PREFIX] [-d FILES_DIRECTORY] [-l LISTEN_ADDRESS] [-c CONNECT_ADDRESS]"
    echo "  -p PREFIX           Set the username prefix (default: OPENAIGPT)"
    echo "  -d FILES_DIRECTORY  Set the directory containing input text files (required)"
    echo "  -c CONNECT_ADDRESS  Specify the address of the server to connect to (default: 127.0.0.1:4000)"
    exit 1
}

# Parse command line options
while getopts ":p:d:l:c:" opt; do
    case $opt in
        p) USERNAME_PREFIX="$OPTARG" ;;
        d) FILES_DIR="$OPTARG" ;;
        l) LISTEN_ADDR="$OPTARG" ;;
        c) CONNECT_ADDR="$OPTARG" ;;
        \?) echo "Invalid option -$OPTARG" >&2; usage ;;
    esac
done

# Check if FILES_DIR is provided and exists
if [ -z "$FILES_DIR" ] || [ ! -d "$FILES_DIR" ]; then
    echo "Error: You must provide a valid directory containing input files"
    usage
fi

echo "Reading files from: $FILES_DIR"

# Get all text files from the directory (non-recursive, handle spaces)
files=()
while IFS= read -r -d '' file; do
    files+=("$file")
done < <(find "$FILES_DIR" -maxdepth 1 -type f -name "*.txt" -print0)

file_count=${#files[@]}
if [ $file_count -eq 0 ]; then
    echo "Error: No text files found in $FILES_DIR"
    exit 1
fi

echo "Found $file_count text files. Starting instances..."

# Run the Go program for each file
for ((i=0; i<file_count; i++)); do
    username="${USERNAME_PREFIX}${i}"
    file_path="${files[$i]}"
    
    echo "Starting instance $((i+1)) with username: $username, file: $file_path"
    
    # Read file content
    file_content=$(<"$file_path")
    
    # Escape special characters in the file content
    escaped_content=$(printf '%q' "$file_content")
    
    # Run the Go program with the specified flags
    exec go run agents/chat.go -username "$username" -input "$escaped_content" -connect "$CONNECT_ADDR" &
done

# Wait for all background processes to finish
wait

echo "All instances have been started."
