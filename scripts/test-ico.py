from PIL import Image
import os

# Create test images
sizes = [16, 32, 48, 64, 128, 256]
images = []

for size in sizes:
    img = Image.new('RGB', (size, size), (100, 100, 100))
    images.append(img)

# Save with all sizes
images[0].save('test.ico', format='ICO', sizes=[(s, s) for s in sizes], append_images=images[1:])

print(f"ICO size: {os.path.getsize('test.ico')} bytes")

# Check what's in the file
with open('test.ico', 'rb') as f:
    data = f.read()
    print(f"First 6 bytes: {data[:6].hex()}")
