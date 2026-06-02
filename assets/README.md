# Assets

| File | Use |
|------|-----|
| `opwax.ico` | `rsrc` embed (16/32/48/256 in one ICO) |
| `opwax.png` | Fyne `SetIcon` (512×512) |

Optional ICO from PNG:

```powershell
python -c "from PIL import Image; i=Image.open('assets/opwax.png').convert('RGBA'); i.save('assets/opwax.ico', format='ICO', sizes=[(256,256),(48,48),(32,32),(16,16)])"
```

Then `.\scripts\build.bat`.
