import "../css/input.css"
import "htmx.org"
import Alpine from "alpinejs"
import intersect from "@alpinejs/intersect"
import * as m from "motion"

import "./carousel.min.js"
import "./dialog.min.js"
import "./input.min.js"
import "./label.min.js"
import "./popover.min.js"
import "./selectbox.min.js"
import "./textarea.min.js"
import "./sidebar.min.js"

window.Alpine = Alpine
window.m = m
Alpine.plugin(intersect)
Alpine.start()