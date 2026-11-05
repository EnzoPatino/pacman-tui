# Pac-Man TUI: tecnología recomendada e historias de usuario

## Tecnología recomendada: Go + Bubbletea (o tcell)

**Por qué:**

- Binario único: publicar en AUR/GitHub Releases es trivial (`go build` y listo).
- Curva de aprendizaje razonable, bastante más fácil que Rust.
- Consumo de CPU/RAM muy bajo, importante para un screensaver que corre horas.
- Compilación cruzada simple (Linux/macOS).

**De más a menos conveniente:**

1. **Go**: mejor equilibrio. Elegirlo.
2. **Python (curses/blessed)**: solo si se quiere un prototipo en un día; más pesado y más difícil de distribuir.
3. **Rust + Ratatui**: excelente rendimiento y portafolio, pero la curva es más dura si recién se empieza.
4. **C/ncurses**: innecesariamente trabajoso.
5. **Bash**: no usarlo, es lento y frágil.

**Aclaración sobre el documento original:** para el juego alcanza un solo loop con tick (ej. 10 FPS) que actualice todo. No hace falta una goroutine por fantasma. Con tcell directo hay más control del dibujo; Bubbletea es más cómodo pero está pensado para apps, no para animaciones constantes.

## Historias de usuario revisadas y ampliadas

**Problemas del original:** faltan reglas de fantasmas, power pellets, configuración, requisitos no funcionales y casos borde (terminal chica, señales).

### US-1: Ejecución autónoma (ajustada)

**Como** usuario de ricing, **quiero** iniciar una simulación automática con un comando, **para** decorar mi terminal sin interactuar.

1. Inicia al instante sin input.
2. Pac-Man y fantasmas se mueven celda a celda, sin atravesar muros.
3. Al comer todos los puntos, carga un laberinto nuevo.
4. Al ser atrapado, reinicia el ciclo (sin pantalla de game over bloqueante).
5. `q`, `ESC`, `Ctrl+C` y `SIGTERM` cierran limpio: cursor visible, pantalla alternativa restaurada.

### US-2: Adaptación al tamaño (ajustada)

1. Detecta `rows × cols` al iniciar.
2. En `SIGWINCH` regenera/reescala el laberinto sin crashear.
3. Si la terminal es menor al mínimo soportado, muestra un aviso en vez de romperse.
4. Compensa la proporción de celdas (cada celda = 2 columnas para verse cuadrada).

### US-3: IA de Pac-Man

**Como** espectador, **quiero** que Pac-Man juegue de forma convincente, **para** que la animación sea entretenida.

1. Prioriza el punto más cercano (BFS).
2. Evita fantasmas cercanos y no se queda en bucles.
3. Persigue fantasmas asustados cuando hay power pellet activo.

### US-4: Comportamiento de fantasmas

1. Cada fantasma tiene personalidad distinta (perseguidor, emboscador, aleatorio).
2. No giran 180° salvo cambio de modo.
3. Alternan modos persecución/dispersión.
4. Al comer un power pellet pasan a modo asustado y vuelven tras unos segundos.

### US-5: Power pellets y puntaje

1. Hay power pellets en el mapa y son visibles.
2. Comer un fantasma lo devuelve a su base.
3. Puntaje y vidas opcionales en una barra discreta (se pueden ocultar).

### US-6: Laberintos variados

**Quiero** laberintos distintos en cada ciclo, **para** que no se vea repetitivo.

1. Mapas predefinidos y/o generados proceduralmente, siempre con todas las celdas alcanzables.
2. Soporta túneles laterales (wrap-around).

### US-7: Temas de color

1. Paletas incluidas: Catppuccin, Dracula, Nord.
2. Detecta TrueColor; si no hay, cae a 256 o 16 colores.
3. Opción de respetar los colores de la terminal (`--theme terminal`).

### US-8: Configuración por CLI

1. Flags: `--fps`, `--theme`, `--ghosts N`, `--no-hud`, `--seed`, `--version`, `--help`.
2. Archivo opcional `~/.config/pacman-tui/config.toml`.

### US-9: Rendimiento y estabilidad (no funcional)

1. CPU < 2 % en idle a 10 FPS en hardware moderno.
2. Sin parpadeo (diff de pantalla, no redibujar todo).
3. Sin crecimiento de memoria tras 24 h.

### US-10: Instalación y distribución

1. Se instala con un solo comando (`go install` / AUR / release binario).
2. README con GIF demo e instrucciones.
3. Compatible con terminales comunes (kitty, alacritty, foot, wezterm, tmux).

### US-11: Uso como screensaver real (opcional)

1. Modo `--screensaver`: sale con cualquier tecla.
2. Integrable con `hypridle`/`xscreensaver` o lanzable en una terminal fullscreen.

## Prioridad MVP

US-1, US-2, US-3, US-4 (básico), US-6 (un mapa). Lo demás va después.
