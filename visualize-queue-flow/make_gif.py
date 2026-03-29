#!/usr/bin/env python3
"""Assemble a VHS PNG frame directory into an animated GIF using Pillow."""
import sys
import os
import glob
from PIL import Image

def make_gif(frames_dir, output_gif, fps=20, every_nth=1):
    """frames_dir is like 'queue_l1.png/', output_gif is 'queue_l1.gif'."""
    text_frames = sorted(glob.glob(os.path.join(frames_dir, 'frame-text-*.png')))
    cursor_frames = sorted(glob.glob(os.path.join(frames_dir, 'frame-cursor-*.png')))
    
    if not text_frames:
        print(f"No frames found in {frames_dir}", file=sys.stderr)
        return False
    
    # Pair up text and cursor frames
    frames = []
    for i, tf in enumerate(text_frames[::every_nth]):
        img = Image.open(tf).convert('RGBA')
        # Composite cursor on top if available
        if cursor_frames and i < len(cursor_frames[::every_nth]):
            cf = cursor_frames[::every_nth][i]
            cursor = Image.open(cf).convert('RGBA')
            img = Image.alpha_composite(img, cursor)
        img = img.convert('P', dither=Image.Dither.NONE, palette=Image.Palette.ADAPTIVE, colors=256)
        frames.append(img)
    
    if not frames:
        return False
    
    duration_ms = int(1000 / fps * every_nth)
    frames[0].save(
        output_gif,
        save_all=True,
        append_images=frames[1:],
        optimize=False,
        duration=duration_ms,
        loop=0,
    )
    print(f"  → {output_gif}  ({len(frames)} frames @ {fps}fps, every {every_nth}th)")
    return True

if __name__ == '__main__':
    import argparse
    p = argparse.ArgumentParser()
    p.add_argument('frames_dir')
    p.add_argument('output_gif')
    p.add_argument('--fps', type=int, default=20)
    p.add_argument('--every', type=int, default=1, help='take every Nth frame (reduce size)')
    args = p.parse_args()
    make_gif(args.frames_dir, args.output_gif, args.fps, args.every)
