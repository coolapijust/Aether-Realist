#!/usr/bin/env python3
"""Generate Tauri app icons using Pillow."""

from pathlib import Path
from PIL import Image, ImageDraw

def create_icon(size):
    """Create a simple gradient icon."""
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    # Corner radius (25% of size)
    radius = size // 4
    
    # Draw rounded rectangle background with gradient
    for y in range(size):
        # Blue to purple gradient
        ratio = y / size
        r = int(59 + (139 - 59) * ratio)   # 3b82f6 to 8b5cf6
        g = int(130 + (92 - 130) * ratio)
        b = int(246 + (246 - 246) * ratio)
        draw.line([(0, y), (size, y)], fill=(r, g, b, 255))
    
    # Create mask for rounded corners
    mask = Image.new('L', (size, size), 0)
    mask_draw = ImageDraw.Draw(mask)
    mask_draw.rounded_rectangle([0, 0, size, size], radius=radius, fill=255)
    
    # Apply mask
    img.putalpha(mask)
    
    # Draw white circle in center (outer ring)
    center = size // 2
    outer_r = int(size * 0.275)  # 140/512 ≈ 0.275
    inner_r = int(size * 0.117)  # 60/512 ≈ 0.117
    line_width = max(2, size // 16)  # 32/512 = 0.0625
    
    draw.ellipse(
        [center - outer_r, center - outer_r, center + outer_r, center + outer_r],
        outline=(255, 255, 255, 255),
        width=line_width
    )
    
    # Draw inner filled circle
    draw.ellipse(
        [center - inner_r, center - inner_r, center + inner_r, center + inner_r],
        fill=(255, 255, 255, 255)
    )
    
    # Draw cross lines
    line_len = int(size * 0.058)  # 30/512 ≈ 0.058
    
    # Top
    draw.line(
        [center, center - outer_r - line_len, center, center - outer_r + line_len],
        fill=(255, 255, 255, 255),
        width=line_width
    )
    # Bottom
    draw.line(
        [center, center + outer_r - line_len, center, center + outer_r + line_len],
        fill=(255, 255, 255, 255),
        width=line_width
    )
    # Left
    draw.line(
        [center - outer_r - line_len, center, center - outer_r + line_len, center],
        fill=(255, 255, 255, 255),
        width=line_width
    )
    # Right
    draw.line(
        [center + outer_r - line_len, center, center + outer_r + line_len, center],
        fill=(255, 255, 255, 255),
        width=line_width
    )
    
    return img

def generate_icons():
    """Generate all required icon files."""
    base_dir = Path(__file__).parent.parent / "gui" / "src-tauri" / "icons"
    
    # Generate PNGs
    sizes = [32, 128, 256, 512]
    for size in sizes:
        if size == 256:
            output = base_dir / "icon.png"
        else:
            output = base_dir / f"{size}x{size}.png"
        
        print(f"Generating {output.name}...")
        img = create_icon(size)
        img.save(output, 'PNG')
    
    # Generate 128x128@2x.png (256x256 for retina)
    print("Generating 128x128@2x.png...")
    img = create_icon(256)
    img.save(base_dir / "128x128@2x.png", 'PNG')
    
    # Generate ICO for Windows (multi-resolution)
    print("Generating icon.ico...")
    ico_sizes = [16, 32, 48, 64, 128, 256]
    ico_images = [create_icon(s).convert('RGBA') for s in ico_sizes]
    ico_images[0].save(
        base_dir / "icon.ico",
        format='ICO',
        sizes=[(s, s) for s in ico_sizes],
        append_images=ico_images[1:]
    )
    
    # Generate ICNS for macOS (as PNG since ICNS needs macOS tools)
    print("Generating icon.icns placeholder...")
    img = create_icon(512)
    img.save(base_dir / "icon.icns.png", 'PNG')
    
    print("\nAll icons generated!")

if __name__ == "__main__":
    try:
        from PIL import Image, ImageDraw
    except ImportError:
        print("Installing Pillow...")
        import subprocess
        import sys
        subprocess.check_call([sys.executable, "-m", "pip", "install", "Pillow", "-q"])
        from PIL import Image, ImageDraw
    
    generate_icons()
