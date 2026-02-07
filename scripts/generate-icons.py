#!/usr/bin/env python3
"""Generate Tauri app icons from SVG source."""

import subprocess
import sys
from pathlib import Path

# Icon sizes needed for Tauri
SIZES = [32, 128, 256, 512]

def check_cairosvg():
    """Check if cairosvg is available."""
    try:
        import cairosvg
        return True
    except ImportError:
        return False

def install_cairosvg():
    """Install cairosvg using pip."""
    print("Installing cairosvg...")
    subprocess.check_call([sys.executable, "-m", "pip", "install", "cairosvg", "-q"])

def generate_icons():
    """Generate PNG icons from SVG."""
    import cairosvg
    
    base_dir = Path(__file__).parent.parent / "gui" / "src-tauri" / "icons"
    svg_path = base_dir / "icon.svg"
    
    if not svg_path.exists():
        print(f"Error: {svg_path} not found!")
        sys.exit(1)
    
    # Generate PNGs
    for size in SIZES:
        output = base_dir / f"{size}x{size}.png"
        if size == 256:
            output = base_dir / "icon.png"
        
        print(f"Generating {output.name}...")
        cairosvg.svg2png(
            url=str(svg_path),
            write_to=str(output),
            output_width=size,
            output_height=size
        )
    
    # Generate 128x128@2x.png (256x256 for retina)
    print("Generating 128x128@2x.png...")
    cairosvg.svg2png(
        url=str(svg_path),
        write_to=str(base_dir / "128x128@2x.png"),
        output_width=256,
        output_height=256
    )
    
    # Generate ICO for Windows
    print("Generating icon.ico...")
    try:
        from PIL import Image
        # Create multi-resolution ICO
        sizes = [(16, 16), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]
        images = []
        
        for w, h in sizes:
            # Generate PNG at this size
            png_data = cairosvg.svg2png(
                url=str(svg_path),
                output_width=w,
                output_height=h
            )
            from io import BytesIO
            img = Image.open(BytesIO(png_data))
            images.append(img)
        
        images[0].save(
            base_dir / "icon.ico",
            format='ICO',
            sizes=[(img.width, img.height) for img in images],
            append_images=images[1:]
        )
        print("icon.ico generated successfully!")
    except ImportError:
        print("PIL not available, skipping ICO generation")
    
    # Generate ICNS for macOS
    print("Generating icon.icns...")
    try:
        from PIL import Image
        import tempfile
        import shutil
        
        base_dir = Path(__file__).parent.parent / "gui" / "src-tauri" / "icons"
        
        # Create iconset directory
        iconset_dir = tempfile.mkdtemp(suffix='.iconset')
        
        # macOS icon sizes
        mac_sizes = [
            (16, '16x16'),
            (32, '16x16@2x'),
            (32, '32x32'),
            (64, '32x32@2x'),
            (128, '128x128'),
            (256, '128x128@2x'),
            (256, '256x256'),
            (512, '256x256@2x'),
            (512, '512x512'),
            (1024, '512x512@2x'),
        ]
        
        for size, name in mac_sizes:
            png_data = cairosvg.svg2png(
                url=str(base_dir / "icon.svg"),
                output_width=size,
                output_height=size
            )
            from io import BytesIO
            img = Image.open(BytesIO(png_data))
            img.save(f"{iconset_dir}/icon_{name}.png")
        
        # Use iconutil to compile (macOS only) or create manually
        # For cross-platform, we'll just create a 512x512 PNG as icon.icns alternative
        png_data = cairosvg.svg2png(
            url=str(base_dir / "icon.svg"),
            output_width=512,
            output_height=512
        )
        from io import BytesIO
        img = Image.open(BytesIO(png_data))
        img.save(base_dir / "icon.icns.png")
        print("icon.icns.png generated (placeholder for macOS)")
        
        # Clean up
        shutil.rmtree(iconset_dir)
        
    except Exception as e:
        print(f"ICNS generation skipped: {e}")
    
    print("\nAll icons generated!")

if __name__ == "__main__":
    if not check_cairosvg():
        install_cairosvg()
    
    generate_icons()
