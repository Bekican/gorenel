from pathlib import Path

from PIL import Image


def main() -> None:
    src = Path("c:/Users/Bekir Can/Desktop/advancedbackend/gorenel/web-dashboard/public/logo.png")
    img = Image.open(src).convert("RGB")

    # Lightweight app/logo asset used in nav and cards.
    logo_square = img.resize((256, 256), Image.Resampling.LANCZOS)
    logo_square.save(src, format="PNG", optimize=True)

    # Social sharing image for OpenGraph/Twitter cards.
    og = img.resize((1200, 630), Image.Resampling.LANCZOS)
    og.save(src.parent / "og-cover.jpg", format="JPEG", quality=82, optimize=True, progressive=True)

    print("optimized logo and generated og-cover.jpg")


if __name__ == "__main__":
    main()
