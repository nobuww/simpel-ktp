import "../css/input.css"
import "htmx.org"
import Alpine from "alpinejs"
import intersect from "@alpinejs/intersect"
import * as m from "motion"

window.Alpine = Alpine
window.m = m
Alpine.plugin(intersect)
Alpine.start()