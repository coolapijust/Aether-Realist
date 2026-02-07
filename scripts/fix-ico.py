#!/usr/bin/env python3
"""Fix the icon.ico file - ensure it has proper sizes."""

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

def main():
    base_dir = Path(__file__).parent.parent / "gui" / "src-tauri" / "icons"
    
    # Generate individual sizes
    sizes = [16, 32, 48, 64, 128, 256]
    images = []
    
    for size in sizes:
        img = create_icon(size)
        # Convert to RGB for ICO (some Windows versions prefer this)
        rgb_img = Image.new('RGB', (size, size), (255, 255, 255))
        rgb_img.paste(img, mask=img.split()[3])  # Use alpha as mask
        images.append(rgb_img)
        print(f"Created {size}x{size}")
    
    # Save as ICO with all sizes
    ico_path = base_dir / "icon.ico"
    images[0].save(
        ico_path,
        format='ICO',
        sizes=[(s, s) for s in sizes],
        append_images=images[1:]
    )
    
    print(f"Saved {ico_path} ({ico_path.stat().st_size} bytes)")

if __name__ == "__main__":
    try:
        from PIL import Image, ImageDraw
    except ImportError:
        import subprocess
        import sys
        subprocess.check_call([sys.executable, "-m", "pip", "install", "Pillow", "-q"])
        from PIL import Image, ImageDraw
    
    main()
