<div align="center">

# 🟡 PAC-MAN TUI

### El clásico laberinto arcade, jugando solo en tu terminal.

<p>
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go 1.24 o superior">
  <img src="https://img.shields.io/badge/render-tcell-3155C6?style=flat-square" alt="Interfaz de terminal con tcell">
  <img src="https://img.shields.io/badge/modo-autónomo-FFD400?style=flat-square&labelColor=171717" alt="Juego autónomo">
</p>

</div>

Pac-Man recorre laberintos generados en cada partida, recoge puntos y escapa de
tres fantasmas con comportamientos distintos. Las píldoras de poder cambian la
cacería: los fantasmas huyen y Pac-Man puede atraparlos.

## Vista de terminal

```text
                         P A C - M A N
              1UP  001850     HI  001850     LIVES  ● ● ●

      ╭━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╮
      ┃ ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  ┃
      ┃ ·  ╭━━━━╮ ·  ·  ╭━━━━━━━━╮ ·  ·  ·  ·  ·  ┃
      ┃ ·  ╰━━━━╯ ·  ·  ╰━━━━━━━━╯ ·  ·  ·  ·  ·  ┃
      ┃ ·  ·  ·  ·  ·  ·  ◕  ·  ·  ·  ·  ·  ·  ·  ┃
      ┃ ·  ·  ·  ·  ·  ·  👻  ·  ·  ·  ·  ·  ·  ·  ┃
      ╰━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╯

                    PAC-MAN  ·  q / esc para salir
```

*Esquema ilustrativo: cada partida genera un laberinto diferente.*

## Empezar

Necesitas Go 1.24 o superior y una terminal con soporte de color.

```bash
git clone https://github.com/EnzoPatino/pacman-tui.git
cd pacman-tui
go run ./cmd/glimmer
```

También puedes compilar un binario local:

```bash
go build -o pacman-tui ./cmd/glimmer
./pacman-tui
```

Pulsa `q`, `Esc` o `Ctrl+C` para salir. La partida es una animación autónoma;
no hace falta controlar a Pac-Man.

## Qué incluye

- **Laberintos nuevos** en cada ronda, con túneles laterales.
- **Tres fantasmas** que persiguen, anticipan el recorrido o deambulan.
- **Píldoras de poder** para cambiar el sentido de la cacería.
- **Puntaje, vidas y niveles** en una interfaz inspirada en el arcade.
- **Ajuste al tamaño de la terminal** y redibujado al cambiar sus dimensiones.
- **Dibujo incremental** para actualizar solo las celdas que cambiaron.

## Proyecto

```text
cmd/glimmer/       Entrada del programa
internal/game/     Simulación, movimiento y decisiones
internal/maze/     Generación y recorrido del laberinto
internal/ui/       Interfaz, sprites y renderizado
```

## Tecnología

Hecho en Go con [tcell](https://github.com/gdamore/tcell), una biblioteca para
interfaces de terminal. No requiere navegador ni recursos gráficos externos.

---

<div align="center">
  <sub>Un pequeño homenaje al arcade que convirtió un laberinto en una persecución.</sub>
</div>
