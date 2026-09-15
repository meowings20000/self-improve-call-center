const assert = require("node:assert/strict");
const fs = require("node:fs");
const vm = require("node:vm");

const html = fs.readFileSync(new URL("./index.html", `file://${__dirname}/`), "utf8");
const js = fs.readFileSync(new URL("./app.js", `file://${__dirname}/`), "utf8");
const css = fs.readFileSync(new URL("./styles.css", `file://${__dirname}/`), "utf8");
const compose = fs.readFileSync(new URL("../docker-compose.yml", `file://${__dirname}/`), "utf8");
const nginx = fs.readFileSync(new URL("./nginx.conf", `file://${__dirname}/`), "utf8");

assert.doesNotThrow(() => new vm.Script(js), "app.js should parse");
for (const id of ["choices", "checkButton", "feedback", "skillGrid", "lessonProgress"]) {
  assert.match(html, new RegExp(`id=["']${id}["']`), `missing required UI element #${id}`);
}
assert.match(js, /\/api\/sessions/, "frontend should start an adaptive session");
assert.match(js, /exerciseId/, "frontend should submit an exercise answer");
assert.match(js, /URLSearchParams/, "local development API override should be explicit");
assert.doesNotMatch(js, /window\.location\.port\s*===/, "Docker must use the same-origin Nginx API proxy by default");
assert.doesNotMatch(html, /https?:\/\//, "runtime assets should not require the network");
assert.match(compose, /127\.0\.0\.1:8080:8080/, "backend port must bind only to loopback");
assert.match(compose, /127\.0\.0\.1:3000:80/, "frontend port must bind only to loopback");
assert.match(nginx, /proxy_set_header Origin \$http_origin;/, "proxy must preserve the browser Origin header for backend CORS checks");
assert.doesNotMatch(nginx, /proxy_set_header Origin "";/, "proxy must not erase Origin and bypass backend CORS checks");

// Responsive navigation: below 800px the sidebar is hidden, so the menu button must open it
// instead of jumping straight to the dashboard and stranding the learner.
assert.match(js, /aria-expanded/, "menu button should expose its expanded state");
assert.match(js, /nav-open/, "menu button should toggle the mobile navigation panel");
assert.doesNotMatch(
  js,
  /menuButton"\)\.addEventListener\("click", \(\) => showView\("progress"\)\)/,
  "menu button must open navigation rather than jump to the skill map"
);
assert.match(
  css,
  /@media\(max-width:800px\)\{[\s\S]*\.app-shell\.nav-open \.sidebar\{[^}]*display:(flex|block)/,
  "mobile navigation must be reachable inside the small-viewport media query"
);
console.log("frontend checks passed");
