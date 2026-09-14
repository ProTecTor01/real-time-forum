const assert = require("node:assert/strict");
const { readFileSync } = require("node:fs");
const { join } = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");

// Run the actual SPA handlers with a small DOM and deterministic timers.
function fixture(options = {}) {
  let now = 10000;
  let nextTimer = 0;
  const timers = new Map();
  const elements = new Map();
  function element(id) {
    if (!elements.has(id)) {
      const classes = new Set(id === "typing-indicator" ? ["hidden"] : []);
      const listeners = {};
      elements.set(id, {
        value: "", textContent: "", innerHTML: "", style: {},
        children: [],
        appendChild(child) { this.children.push(child); },
        remove() {},
        classList: { add: c => classes.add(c), remove: c => classes.delete(c), contains: c => classes.has(c) },
        addEventListener: (type, callback) => { (listeners[type] ??= []).push(callback); },
        dispatch: (type, event = {}) => { for (const callback of listeners[type] || []) callback(event); },
      });
    }
    return elements.get(id);
  }
  const document = Object.assign(element("document"), {
    hidden: false, body: element("body"),
    getElementById: element,
    querySelector: () => element("message-input"),
    querySelectorAll: () => [],
    createElement: () => element(`created-${elements.size}`),
  });
  const window = element("window");
  class Socket {
    static OPEN = 1;
    readyState = 1;
    sent = [];
    send(raw) { this.sent.push(JSON.parse(raw)); }
    close() { this.readyState = 3; }
  }
  const context = vm.createContext({
    document, window, WebSocket: Socket, console,
    location: { protocol: "http:", host: "localhost" },
    Date: { now: () => now },
    requestAnimationFrame: callback => callback(),
    fetch: async () => ({ ok: true, json: async () => options.loadUser ?? { id: 1, username: "Alice" } }),
    alert: () => {},
    setTimeout: (callback, delay) => { const id = ++nextTimer; timers.set(id, { callback, at: now + delay }); return id; },
    clearTimeout: id => timers.delete(id),
  });
  const source = readFileSync(join(__dirname, "app.js"), "utf8").replace(/init\(\)\.catch\(err => \{\s*console\.error\(err\);\s*\}\);/, "");
  vm.runInContext(source + "\nglobalThis.app = { state, showView, sendTypingState, setTypingIndicator, connectWS, applyPresence, bindUI };", context);
  const app = context.app;
  app.bindUI();
  app.state.activeChat = { id: 2, username: "Bob" };
  app.showView("chat");
  app.connectWS();
  const input = element("message-input");
  return {
    ...app, input, document, window,
    toast: element("toast-root"),
    indicator: element("typing-indicator"), name: element("typing-name"), form: element("message-form"),
    advance(ms) {
      const end = now + ms;
      while (true) {
        const pending = [...timers].filter(([, timer]) => timer.at <= end).sort((a, b) => a[1].at - b[1].at)[0];
        if (!pending) break;
        now = pending[1].at;
        timers.delete(pending[0]);
        pending[1].callback();
      }
      now = end;
    },
    receive(typing, from = 2) {
      app.state.ws.onmessage({ data: JSON.stringify({ type: "typing", data: { from_user_id: from, username: "Bob", typing } }) });
    },
    type(value = "Hello") { input.value = value; input.dispatch("input"); },
  };
}

test("continuous input refreshes typing without hiding the indicator", () => {
  const sender = fixture();
  const receiver = fixture();
  let delivered = 0;
  for (let i = 0; i < 100; i++) {
    sender.type("a".repeat(i + 1));
    const events = sender.state.ws.sent.slice(delivered);
    delivered += events.length;
    for (const event of events) receiver.receive(event.data.typing);
    assert.equal(receiver.indicator.classList.contains("hidden"), false);
    sender.advance(100);
    receiver.advance(100);
  }
  assert.equal(delivered, 10, "send at most one refresh per second");
  assert.equal(receiver.name.textContent, "Bob ");
  sender.advance(1200);
  assert.equal(sender.state.ws.sent.at(-1).data.typing, false);
  receiver.receive(false);
  assert.equal(receiver.indicator.classList.contains("hidden"), true);
});

test("empty input, focus loss, backgrounding, navigation and submit stop typing", () => {
  const stops = [
    f => f.type(""),
    f => f.type("   "),
    f => f.input.dispatch("blur"),
    f => f.window.dispatch("blur"),
    f => f.window.dispatch("pagehide"),
    f => { f.document.hidden = true; f.document.dispatch("visibilitychange"); },
    f => f.showView("feed"),
    f => f.form.dispatch("submit", { preventDefault() {}, target: { body: f.input } }),
  ];
  for (const stop of stops) {
    const f = fixture();
    f.type();
    stop(f);
    assert.equal(f.state.ws.sent.at(-1).data.typing, false);
    assert.equal(f.state.isTyping, false);
    assert.equal(f.state.typingTimeout, null);
  }
});

test("recipient switching stops the old conversation", () => {
  const f = fixture();
  f.type();
  f.sendTypingState(true, 3);
  assert.equal(f.state.ws.sent[1].data.to_user_id, 2);
  assert.equal(f.state.ws.sent[1].data.typing, false);
  assert.equal(f.state.ws.sent[2].data.to_user_id, 3);
  assert.equal(f.state.ws.sent[2].data.typing, true);
});

test("remote state expires and is cleared by presence and socket closure", async () => {
  const f = fixture();
  f.receive(true, 3);
  assert.equal(f.indicator.classList.contains("hidden"), true);
  f.receive(true);
  f.advance(4000);
  assert.equal(f.indicator.classList.contains("hidden"), true);
  f.receive(true);
  f.applyPresence([]);
  assert.equal(f.indicator.classList.contains("hidden"), true);
  f.receive(true);
  f.state.ws.close();
  await f.state.ws.onclose();
  assert.equal(f.indicator.classList.contains("hidden"), true);
  f.showView("feed");
  f.receive(true);
  assert.equal(f.indicator.classList.contains("hidden"), true);
});

test("an old socket cannot reset a newer connection", async () => {
  const f = fixture();
  const oldSocket = f.state.ws;
  f.connectWS();
  f.type();
  f.receive(true);
  oldSocket.close();
  await oldSocket.onclose();
  assert.equal(f.state.isTyping, true);
  assert.equal(f.indicator.classList.contains("hidden"), false);
  oldSocket.onmessage({ data: JSON.stringify({ type: "typing", data: { from_user_id: 2, typing: false } }) });
  assert.equal(f.indicator.classList.contains("hidden"), false);
});

test("temporary disconnection reconnects and clears typing", async () => {
  const f = fixture();
  f.state.user = { id: 1 };
  f.type();
  f.receive(true);
  const socket = f.state.ws;
  socket.close();
  await socket.onclose({ code: 1006 });
  assert.equal(f.state.isTyping, false);
  assert.equal(f.indicator.classList.contains("hidden"), true);
  f.advance(999);
  assert.equal(f.state.ws, socket);
  f.advance(1);
  assert.notEqual(f.state.ws, socket);
  f.state.ws.onopen();
  assert.equal(f.state.wsReconnectAttempts, 0);
});

test("protocol errors are visible and do not trigger an endless reconnect loop", async () => {
  const f = fixture();
  f.state.user = { id: 1 };
  const socket = f.state.ws;
  socket.close();
  await socket.onclose({ code: 1002, reason: "RSV1 set, bad opcode 7, bad MASK" });
  assert.match(f.toast.children[0].textContent, /Соединение чата прервано/);
  assert.equal(f.state.wsReconnectTimeout, null);
  f.advance(20000);
  assert.equal(f.state.ws, socket);
});

test("session rejection cancels reconnection and shows authentication", async () => {
  const f = fixture();
  f.state.user = { id: 1 };
  const socket = f.state.ws;
  socket.close();
  await socket.onclose({ code: 1008 });
  assert.equal(f.state.user, null);
  assert.equal(f.state.ws, null);
  assert.equal(f.state.wsReconnectTimeout, null);
});

test("failed reconnects are bounded and back off", async () => {
  const f = fixture();
  f.state.user = { id: 1 };
  for (const delay of [1000, 2000, 4000, 8000, 10000]) {
    const socket = f.state.ws;
    socket.close();
    await socket.onclose({ code: 1006 });
    f.advance(delay - 1);
    assert.equal(f.state.ws, socket);
    f.advance(1);
    assert.notEqual(f.state.ws, socket);
  }
  const socket = f.state.ws;
  socket.close();
  await socket.onclose({ code: 1006 });
  assert.equal(f.state.wsReconnectTimeout, null);
  assert.match(f.toast.children[0].textContent, /Не удалось восстановить/);
});
