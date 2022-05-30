package ui

import (
	"image/color"
	"poa-manager/res"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type MyTheme struct{}

func (m *MyTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameBackground {
		if variant == theme.VariantLight {
			return color.White
		}
		return color.Black
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (m *MyTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	if name == theme.IconNameHome {
		fyne.NewStaticResource("myHome", res.Ic_main.StaticContent)
	}

	return theme.DefaultTheme().Icon(name)
}

// TextFont returns the font resource for the regular font style
func (t *MyTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return res.NanumBarunGothic
	}
	if style.Bold {
		if style.Italic {
			return res.NanumBarunGothic
		}
		return res.NanumBarunGothic
	}
	if style.Italic {
		return res.NanumBarunGothic
	}
	return res.NanumBarunGothic
}

func (t *MyTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
