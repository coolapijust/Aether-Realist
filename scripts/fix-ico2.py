#!/usr/bin/env python3
"""Fix the icon.ico file using a different approach."""

from pathlib import Path
from PIL import Image

def create_icon(size):
    """Create a simple gradient icon."""
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    
    # Create gradient background
    pixels = img.load()
    for y in range(size):
        for x in range(size):
            ratio = y / size
            r = int(59 + (139 - 59) * ratio)
            g = int(130 + (92 - 130) * ratio)
            b = int(246 + (246 - 246) * ratio)
            pixels[x, y] = (r, g, b, 255)
    
    # Create mask for rounded corners
    mask = Image.new('L', (size, size), 0)
    from PIL import ImageDraw
    mask_draw = ImageDraw.Draw(mask)
    radius = size // 4
    mask_draw.rounded_rectangle([0, 0, size-1, size-1], radius=radius, fill=255)
    
    # Apply mask
    img.putalpha(mask)
    
    # Draw white elements
    draw = ImageDraw.Draw(img)
    center = size // 2
    outer_r = int(size * 0.275)
    inner_r = int(size * 0.117)
    line_width = max(2, size // 16)
    
    # Outer ring
    draw.ellipse(
        [center - outer_r, center - outer_r, center + outer_r, center + outer_r],
        outline=(255, 255, 255, 255),
        width=line_width
    )
    
    # Inner circle
    draw.ellipse(
        [center - inner_r, center - inner_r, center + inner_r, center + inner_r],
        fill=(255, 255, 255, 255)
    )
    
    # Cross lines
    line_ext = int(size * 0.058)
    for dx, dy in [(0, -1), (0, 1), (-1, 0), (1, 0)]:
        x1 = center + dx * (outer_r - line_ext)
        y1 = center + dy * (outer_r - line_ext)
        x2 = center + dx * (outer_r + line_ext)
        y2 = center + dy * (outer_r + line_ext)
        draw.line([(x1, y1), (x2, y2)], fill=(255, 255, 255, 255), width=line_width)
    
    return img

def create_ico_file(output_path, sizes=[16, 32, 48, 64, 128, 256]):
    """Create a proper multi-resolution ICO file."""
    
    # ICO file header
    # Reserved (2 bytes) + Type (2 bytes) + Count (2 bytes)
    header = bytes([
        0x00, 0x00,  # Reserved
        0x01, 0x00,  # Type: 1 = ICO
        len(sizes) & 0xFF, (len(sizes) >> 8) & 0xFF,  # Count
    ])
    
    # ICONDIRENTRY for each icon (16 bytes each)
    entries = []
    image_data_offset = 6 + len(sizes) * 16  # Header + entries
    image_data = []
    
    for size in sizes:
        # Create the image
        img = create_icon(size)
        
        # Convert to BMP format for ICO (ICO uses BMP without file header)
        # For 32-bit images with alpha, we need BITMAPV4HEADER
        
        # Save to memory as PNG first (easier to handle alpha)
        from io import BytesIO
        png_buffer = BytesIO()
        img.save(png_buffer, format='PNG')
        png_data = png_buffer.getvalue()
        
        # ICO directory entry
        # Width, Height, Colors, Reserved, Planes, BitCount, Size, Offset
        entry = bytes([
            size if size < 256 else 0,  # Width (0 means 256)
            size if size < 256 else 0,  # Height (0 means 256)
            0,  # Colors (0 = >256)
            0,  # Reserved
            0x01, 0x00,  # Planes
            0x20, 0x00,  # BitCount (32 bits)
        ])
        # Size (4 bytes, little endian)
        entry += len(png_data).to_bytes(4, 'little')
        # Offset (4 bytes, little endian)
        entry += image_data_offset.to_bytes(4, 'little')
        
        entries.append(entry)
        image_data.append(png_data)
        image_data_offset += len(png_data)
    
    # Write ICO file
    with open(output_path, 'wb') as f:
        f.write(header)
        for entry in entries:
            f.write(entry)
        for data in image_data:
            f.write(data)
    
    print(f"Created {output_path} ({output_path.stat().st_size} bytes)")
    print(f"Contains {len(sizes)} images: {sizes}")

def main():
    base_dir = Path(__file__).parent.parent / "gui" / "src-tauri" / "icons"
    create_ico_file(base_dir / "icon.ico")

if __name__ == "__main__":
    try:
        from PIL import Image, ImageDraw
    except ImportError:
        import subprocess
        import sys
        subprocess.check_call([sys.executable, "-m", "pip", "install", "Pillow", "-q"])
        from PIL import Image, ImageDraw
    
    main()
