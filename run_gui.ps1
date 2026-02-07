# Build aetherd backend
echo "Building aetherd..."
go build -ldflags="-s -w" -o gui/src-tauri/bin/aetherd.exe ./cmd/aetherd

if ($LASTEXITCODE -ne 0) {
    echo "Build failed!"
    exit 1
}

echo "aetherd built successfully."

# Run GUI dev mode
cd gui
npm run tauri dev
