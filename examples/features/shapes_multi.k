// Multiple shapes in one file + using Shapes.Color via `as`.

let Shapes = import "examples/features/shapes/Palette.shape"

let color = { name: "red", value: "#ff0000", extra: "drop", }
let one = color as Shapes.Color

let many = [
    color,
    { name: "green", value: "#00ff00", },
] as Shapes.Colors

log("one:", one)
log("many:", many)
