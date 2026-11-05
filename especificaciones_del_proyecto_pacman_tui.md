# Documento de Especificación del Proyecto: Pac-Man TUI / Terminal Screensaver

---

## 1. Nombres Sugeridos para el Repositorio

A continuación se presentan distintas opciones de nombres estructuradas según el enfoque que se le desee dar al repositorio en GitHub:

### Opción A: Estándar y Profesional (Recomendado)
* **`pacman-tui`**: Claro, directo y alineado con las convenciones del ecosistema Linux/Open Source.
* **`tty-pacman`**: Hace referencia directa a las terminales Unix/Linux (`tty`).

### Opción B: Estilo Screensaver / Terminal Artwork
* **`pacman-matrix`**: Ideal si el enfoque visual busca recordar a proyectos clásicos como `cmatrix`.
* **`pacman-pipes`** / **`pacman-ss`**: Guiño directo a utilidades como `pipes.sh`.

### Opción C: Creativo / Marca Propia
* **`ghost-chaser-tui`**: Nombre distintivo que evita conflictos directos con marcas registradas.
* **`auto-pacman`**: Enfatiza la naturaleza 100% autónoma y automatizada de la aplicación.

---

## 2. Historias de Usuario (User Stories)

### Historia de Usuario 1: Ejecución Autónoma (Modo Screensaver)
* **Como** usuario entusiasta de Linux y la personalización de entornos (*ricing*),
* **Quiero** ejecutar un comando en la terminal que inicie una simulación de Pac-Man 100% automática,
* **Para** tener una animación estética e interactiva mientras mantengo la terminal inactiva o decorando mi escritorio.

#### Criterios de Aceptación:
1. La animación debe iniciarse inmediatamente tras invocar el comando, sin requerir interacción inicial del usuario.
2. Pac-Man y los fantasmas deben desplazarse de forma fluida respetando las paredes y los límites de la matriz del laberinto.
3. Al completarse la recolección de los puntos del mapa o al ocurrir una colisión, la partida debe reiniciar automáticamente un nuevo ciclo o laberinto.
4. Presionar teclas de interrupción estándar (`q`, `ESC` o `Ctrl+C`) debe finalizar la aplicación limpiamente, restaurando la visibilidad del cursor y el estado original de la terminal.

---

### Historia de Usuario 2: Adaptabilidad al Tamaño de Pantalla
* **Como** usuario con distintas configuraciones de monitores y tamaños de ventana,
* **Quiero** que el laberinto y los personajes se adapten proporcionalmente al ancho y alto de mi terminal,
* **Para** evitar distorsiones visuales, desbordamientos o caracteres cortados.

#### Criterios de Aceptación:
1. El programa detecta automáticamente las filas y columnas disponibles (`rows` × `cols`) al iniciar.
2. Si la terminal cambia de tamaño durante la ejecución, la aplicación recalcula la grilla o reajusta el lienzo sin romperse.

---

## 3. Matriz Comparativa de Tecnologías

| Tecnología | Facilidad de Desarrollo | Rendimiento y Consumo | Portabilidad / Empaquetado | Recomendación General |
| :--- | :--- | :--- | :--- | :--- |
| **Go (Golang)** + `Bubbletea` / `Tcell` | ⭐⭐⭐⭐ (Media-Alta) | ⭐⭐⭐⭐⭐ (Excelente) | ⭐⭐⭐⭐⭐ (Binario único) | **Opción Principal Recomendada** |
| **Python** + `curses` / `Rich` | ⭐⭐⭐⭐⭐ (Alta) | ⭐⭐⭐ (Medio) | ⭐⭐⭐⭐ (Alta) | **Ideal para Prototipo Rápido** |
| **Rust** + `Ratatui` / `Crossterm` | ⭐⭐⭐ (Media) | ⭐⭐⭐⭐⭐ (Ultra alto) | ⭐⭐⭐⭐⭐ (Binario único) | **Para Máximo Rendimiento** |
| **C / C++** + `ncurses` | ⭐⭐ (Baja) | ⭐⭐⭐⭐⭐ (Ultra alto) | ⭐⭐⭐⭐⭐ (Nativa) | **Enfoque Clásico / Tradicional** |
| **Bash Shell Scripting** | ⭐⭐ (Complejo) | ⭐⭐ (Lento) | ⭐⭐⭐⭐⭐ (Universal) | **Solo Educativo** |

---

## 4. Análisis Detallado y Recomendaciones Tecnológicas

### 1. Go (Golang) — *Recomendación Principal*
* **Librerías clave:** `github.com/charmbracelet/bubbletea`, `github.com/gdamore/tcell`
* **Por qué elegirlo:** Go produce un binario compilado único, estático y ultra liviano que no requiere runtime ni librerías adicionales instaladas en el sistema del usuario. Esto facilita enormemente la creación de paquetes para el Arch User Repository (AUR). Además, las *goroutines* de Go hacen que controlar los hilos de movimiento del Pac-Man y de los fantasmas en paralelo sea sumamente sencillo y eficiente.

### 2. Python — *Recomendación para Prototipado Rápido*
* **Librerías clave:** `curses` (módulo nativo), `blessed`, `rich`
* **Por qué elegirlo:** Es la opción más rápida para escribir la lógica inicial del mapa, la validación de matrices y la Inteligencia Artificial básica en cuestión de horas. Casi todas las distros Linux tienen Python preinstalado.
* **Desventaja:** Su consumo de recursos (CPU/RAM) es superior al de lenguajes compilados, lo que puede ser relevante al tratarse de un salvapantallas de ejecución prolongada.

### 3. Rust — *Recomendación para Portafolio y Rendimiento Extremo*
* **Librerías clave:** `ratatui`, `crossterm`
* **Por qué elegirlo:** Si deseas un proyecto de alto nivel para destacar en tu perfil técnico, Rust ofrece una gestión de memoria perfecta y un rendimiento imbatible con un uso de procesador prácticamente nulo. `Ratatui` es la librería TUI más popular e impresionante del ecosistema actual.

---

## 5. Próximos Pasos Recomendados

1. **Definir la estructura de la grilla:** Representar el mapa mediante una matriz bidimensional de caracteres (`#` para muros, `.` para puntos, `C` para Pac-Man, `G` para fantasmas).
2. **Implementar lógica de movimiento autónomo:** Algoritmo simple de toma de decisiones (por ejemplo, evaluar en cada intersección las casillas contiguas libres y elegir dirección aleatoria o perseguir al objetivo mediante distancia Manhattan).
3. **Estética y Paleta de Colores:** Integrar códigos de escape ANSI / TrueColor compatibles con esquemas populares de personalización de terminales como *Catppuccin*, *Dracula* o *Nord*.