const SVG_MAP_URL = "/assets/floor_1.svg";
const GRAPH_GROUP_ID = "intervuz-graph-overlay";
const ROUTE_GROUP_ID = "intervuz-route-overlay";

const state = {
  rooms: [],
  filteredRooms: [],
  graph: null,
  filteredVertices: [],
  selectedRoomId: "",
  selectedVertexId: "",
  graphVisible: true,
  graphLoaded: false,
  route: null,
  routeStartVertexId: "",
  routeEndVertexId: "",
  svgElement: null,
  originalViewBox: "",
};

const elements = {
  apiBaseInput: document.getElementById("apiBaseInput"),
  apiStatus: document.getElementById("apiStatus"),
  toggleGraphButton: document.getElementById("toggleGraphButton"),
  resetViewButton: document.getElementById("resetViewButton"),
  searchInput: document.getElementById("searchInput"),
  searchSummary: document.getElementById("searchSummary"),
  selectionCard: document.getElementById("selectionCard"),
  vertexCount: document.getElementById("vertexCount"),
  vertexList: document.getElementById("vertexList"),
  roomCount: document.getElementById("roomCount"),
  roomList: document.getElementById("roomList"),
  mapMount: document.getElementById("mapMount"),
  routeBadge: document.getElementById("routeBadge"),
  routeStartLabel: document.getElementById("routeStartLabel"),
  routeEndLabel: document.getElementById("routeEndLabel"),
  accessibleOnlyInput: document.getElementById("accessibleOnlyInput"),
  setStartButton: document.getElementById("setStartButton"),
  setEndButton: document.getElementById("setEndButton"),
  buildRouteButton: document.getElementById("buildRouteButton"),
  swapRouteButton: document.getElementById("swapRouteButton"),
  clearRouteButton: document.getElementById("clearRouteButton"),
  routeStatus: document.getElementById("routeStatus"),
  routeMetrics: document.getElementById("routeMetrics"),
  routeSteps: document.getElementById("routeSteps"),
};

boot().catch((error) => {
  setStatus(`Не удалось запустить фронтенд: ${error.message}`, "error");
});

async function boot() {
  attachUI();
  await loadSvgMap();
  applySearch();
  renderSelection();
  renderRoutePlanner();
  try {
    await loadGraphOverlay();
    setStatus("Карта и граф загружены. Можно выбирать узлы и строить маршрут.", "ok");
  } catch (error) {
    setStatus(error.message, "warning");
  }
}

function attachUI() {
  elements.searchInput.addEventListener("input", () => {
    applySearch();
    renderLists();
  });

  elements.resetViewButton.addEventListener("click", resetViewBox);
  elements.setStartButton.addEventListener("click", () => assignSelectedVertex("start"));
  elements.setEndButton.addEventListener("click", () => assignSelectedVertex("end"));
  elements.buildRouteButton.addEventListener("click", buildRoute);
  elements.swapRouteButton.addEventListener("click", swapRouteVertices);
  elements.clearRouteButton.addEventListener("click", clearRoute);

  elements.toggleGraphButton.addEventListener("click", async () => {
    if (!state.graphLoaded) {
      await loadGraphOverlay();
    }

    state.graphVisible = !state.graphVisible;
    updateGraphVisibility();
  });
}

async function loadSvgMap() {
  const response = await fetch(SVG_MAP_URL);
  if (!response.ok) {
    throw new Error(`SVG не найден по адресу ${SVG_MAP_URL}`);
  }

  const svgText = await response.text();
  const documentFragment = new DOMParser().parseFromString(svgText, "image/svg+xml");
  const svgElement = documentFragment.documentElement;
  if (!svgElement || svgElement.nodeName.toLowerCase() !== "svg") {
    throw new Error("Получен некорректный SVG");
  }

  state.svgElement = svgElement;
  state.originalViewBox = svgElement.getAttribute("viewBox") || "";

  prepareSvg(svgElement);
  elements.mapMount.replaceChildren(svgElement);

  state.rooms = extractRooms(svgElement);
  applySearch();
  renderLists();
}

function prepareSvg(svgElement) {
  svgElement.classList.add("interactive-map");
  svgElement.setAttribute("preserveAspectRatio", "xMidYMid meet");

  const graphGroup = document.createElementNS("http://www.w3.org/2000/svg", "g");
  graphGroup.setAttribute("id", GRAPH_GROUP_ID);

  const routeGroup = document.createElementNS("http://www.w3.org/2000/svg", "g");
  routeGroup.setAttribute("id", ROUTE_GROUP_ID);

  svgElement.append(graphGroup, routeGroup);
}

function extractRooms(svgElement) {
  const candidates = Array.from(
    svgElement.querySelectorAll("path[id], rect[id], polygon[id], polyline[id], ellipse[id], circle[id]")
  );

  const rooms = [];
  for (const element of candidates) {
    const id = element.id.trim();
    if (!id) {
      continue;
    }

    const title = element.querySelector("title")?.textContent?.trim() || humanizeId(id);
    if (!title) {
      continue;
    }

    const room = {
      id,
      title,
      label: [title, id].filter(Boolean).join(" · "),
      element,
    };

    element.classList.add("map-room-shape");
    element.addEventListener("click", (event) => {
      event.preventDefault();
      event.stopPropagation();
      selectRoom(room.id);
    });

    rooms.push(room);
  }

  rooms.sort((left, right) => left.title.localeCompare(right.title, "ru"));
  return rooms;
}

async function loadGraphOverlay() {
  const apiBase = normalizeApiBase(elements.apiBaseInput.value);
  const response = await fetch(`${apiBase}/graph`);
  if (!response.ok) {
    state.graphLoaded = false;
    state.graphVisible = false;
    updateGraphVisibility();
    throw new Error(`Не удалось загрузить граф: ${response.status}`);
  }

  const payload = await response.json();
  state.graph = normalizeGraph(payload);
  state.graphLoaded = true;
  state.graphVisible = true;
  applySearch();
  renderGraph();
  renderLists();
  updateGraphVisibility();
}

function normalizeGraph(payload) {
  const vertices = Array.isArray(payload.vertices)
    ? payload.vertices.map((vertex) => ({
        id: String(vertex.id || ""),
        placeId: String(vertex.placeId || ""),
        building: String(vertex.building || ""),
        floor: Number.parseInt(String(vertex.floor ?? 0), 10) || 0,
        x: Number(vertex.x),
        y: Number(vertex.y),
        isAccessible: Boolean(vertex.isAccessible),
      }))
    : [];

  const edges = Array.isArray(payload.edges)
    ? payload.edges.map((edge) => ({
        id: String(edge.id || ""),
        from: String(edge.from || ""),
        to: String(edge.to || ""),
      }))
    : [];

  return { vertices, edges };
}

function applySearch() {
  const query = elements.searchInput.value.trim().toLowerCase();

  if (!query) {
    state.filteredRooms = [...state.rooms];
    state.filteredVertices = [...(state.graph?.vertices || [])];
  } else {
    state.filteredRooms = state.rooms.filter((room) => room.label.toLowerCase().includes(query));
    state.filteredVertices = (state.graph?.vertices || []).filter((vertex) =>
      vertexSearchText(vertex).includes(query)
    );
  }

  elements.searchSummary.textContent = query
    ? `Найдено помещений: ${state.filteredRooms.length}, узлов: ${state.filteredVertices.length}.`
    : `Показаны все помещения: ${state.filteredRooms.length}, узлы: ${state.filteredVertices.length}.`;
}

function renderLists() {
  renderVertexList();
  renderRoomList();
}

function renderVertexList() {
  const total = state.graph?.vertices?.length || 0;
  elements.vertexCount.textContent = String(total);

  if (state.filteredVertices.length === 0) {
    elements.vertexList.innerHTML = '<p class="caption">Узлы графа не найдены.</p>';
    return;
  }

  const fragment = document.createDocumentFragment();
  for (const vertex of state.filteredVertices) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "room-item";
    if (vertex.id === state.selectedVertexId) {
      button.classList.add("is-active");
    }
    if (vertex.id === state.routeStartVertexId) {
      button.classList.add("is-route-start");
    }
    if (vertex.id === state.routeEndVertexId) {
      button.classList.add("is-route-end");
    }

    const title = document.createElement("span");
    title.className = "room-item-title";
    title.textContent = vertexDisplayName(vertex);

    const subtitle = document.createElement("span");
    subtitle.className = "room-item-subtitle";
    subtitle.textContent = `${vertex.id} · ${formatNumber(vertex.x)}, ${formatNumber(vertex.y)}`;

    button.append(title, subtitle);
    button.addEventListener("click", () => selectVertex(vertex.id));
    fragment.append(button);
  }

  elements.vertexList.replaceChildren(fragment);
}

function renderRoomList() {
  elements.roomCount.textContent = String(state.rooms.length);

  if (state.filteredRooms.length === 0) {
    elements.roomList.innerHTML = '<p class="caption">Помещения не найдены.</p>';
    return;
  }

  const fragment = document.createDocumentFragment();
  for (const room of state.filteredRooms) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "room-item";
    if (room.id === state.selectedRoomId) {
      button.classList.add("is-active");
    }

    const title = document.createElement("span");
    title.className = "room-item-title";
    title.textContent = room.title;

    const subtitle = document.createElement("span");
    subtitle.className = "room-item-subtitle";
    subtitle.textContent = room.id;

    button.append(title, subtitle);
    button.addEventListener("click", () => selectRoom(room.id));
    fragment.append(button);
  }

  elements.roomList.replaceChildren(fragment);
}

function selectRoom(roomID) {
  state.selectedRoomId = roomID;
  state.selectedVertexId = "";
  updateRoomStateClasses();
  renderLists();
  renderSelection();
  updateRoutePlannerButtons();
}

function selectVertex(vertexID) {
  if (!findVertex(vertexID)) {
    return;
  }

  state.selectedVertexId = vertexID;
  state.selectedRoomId = "";
  updateRoomStateClasses();
  renderGraph();
  renderLists();
  renderSelection();
  updateRoutePlannerButtons();
}

function renderSelection() {
  const selectedVertex = getSelectedVertex();
  if (selectedVertex) {
    elements.selectionCard.className = "selection-card";
    elements.selectionCard.innerHTML = `
      <h3 class="selection-title">${escapeHTML(vertexDisplayName(selectedVertex))}</h3>
      <p class="selection-meta">Vertex ID: ${escapeHTML(selectedVertex.id)}</p>
      <p class="selection-meta">Координаты: x=${formatNumber(selectedVertex.x)}, y=${formatNumber(
        selectedVertex.y
      )}</p>
      <p class="selection-meta">Этаж: ${selectedVertex.floor || "?"} · Корпус: ${escapeHTML(
        selectedVertex.building || "?"
      )}</p>
      <span class="selection-chip${selectedVertex.isAccessible ? "" : " is-muted"}">
        ${selectedVertex.isAccessible ? "Доступный узел" : "Недоступный узел"}
      </span>
      <div class="selection-actions">
        <button id="selectionStartButton" class="button button-secondary" type="button">Сделать стартом</button>
        <button id="selectionEndButton" class="button button-secondary" type="button">Сделать финишем</button>
      </div>
    `;

    elements.selectionCard.querySelector("#selectionStartButton")?.addEventListener("click", () => {
      assignVertex("start", selectedVertex.id);
    });
    elements.selectionCard.querySelector("#selectionEndButton")?.addEventListener("click", () => {
      assignVertex("end", selectedVertex.id);
    });
    return;
  }

  const room = getSelectedRoom();
  if (room) {
    const bbox = room.element.getBBox();
    elements.selectionCard.className = "selection-card";
    elements.selectionCard.innerHTML = `
      <h3 class="selection-title">${escapeHTML(room.title)}</h3>
      <p class="selection-meta">SVG ID: ${escapeHTML(room.id)}</p>
      <p class="selection-meta">Область: x=${formatNumber(bbox.x)}, y=${formatNumber(bbox.y)}</p>
      <p class="selection-meta">Размер: ${formatNumber(bbox.width)} × ${formatNumber(bbox.height)}</p>
      <span class="selection-chip is-muted">Маршрут строится по узлам графа, а не по помещениям.</span>
    `;
    return;
  }

  elements.selectionCard.className = "selection-card is-empty";
  elements.selectionCard.innerHTML =
    '<p class="selection-empty">Нажмите на помещение или на узел графа прямо на карте.</p>';
}

function renderGraph() {
  if (!state.svgElement) {
    return;
  }

  const group = state.svgElement.querySelector(`#${GRAPH_GROUP_ID}`);
  if (!group) {
    return;
  }

  group.replaceChildren();
  if (!state.graph) {
    return;
  }

  const verticesByID = new Map(state.graph.vertices.map((vertex) => [vertex.id, vertex]));

  for (const edge of state.graph.edges) {
    const from = verticesByID.get(edge.from);
    const to = verticesByID.get(edge.to);
    if (!from || !to) {
      continue;
    }

    const line = document.createElementNS("http://www.w3.org/2000/svg", "line");
    line.setAttribute("class", "map-graph-edge");
    line.setAttribute("x1", String(from.x));
    line.setAttribute("y1", String(from.y));
    line.setAttribute("x2", String(to.x));
    line.setAttribute("y2", String(to.y));
    group.append(line);
  }

  for (const vertex of state.graph.vertices) {
    const circle = document.createElementNS("http://www.w3.org/2000/svg", "circle");
    circle.setAttribute("class", buildVertexClassName(vertex));
    circle.setAttribute("cx", String(vertex.x));
    circle.setAttribute("cy", String(vertex.y));
    circle.setAttribute("r", "5.5");
    circle.addEventListener("click", (event) => {
      event.preventDefault();
      event.stopPropagation();
      selectVertex(vertex.id);
    });
    group.append(circle);

    if (shouldRenderVertexLabel(vertex)) {
      const label = document.createElementNS("http://www.w3.org/2000/svg", "text");
      label.setAttribute("class", "map-overlay-label");
      label.setAttribute("x", String(vertex.x + 8));
      label.setAttribute("y", String(vertex.y - 8));
      label.textContent = vertexDisplayName(vertex);
      group.append(label);
    }
  }

  updateGraphVisibility();
}

function buildVertexClassName(vertex) {
  const classes = ["map-graph-vertex"];
  if (!vertex.isAccessible) {
    classes.push("is-disabled");
  }
  if (vertex.id === state.selectedVertexId) {
    classes.push("is-selected");
  }
  if (vertex.id === state.routeStartVertexId) {
    classes.push("is-start");
  }
  if (vertex.id === state.routeEndVertexId) {
    classes.push("is-end");
  }
  return classes.join(" ");
}

function shouldRenderVertexLabel(vertex) {
  return (
    vertex.id === state.selectedVertexId ||
    vertex.id === state.routeStartVertexId ||
    vertex.id === state.routeEndVertexId
  );
}

function updateGraphVisibility() {
  const group = state.svgElement?.querySelector(`#${GRAPH_GROUP_ID}`);
  if (group) {
    group.setAttribute("visibility", state.graphVisible ? "visible" : "hidden");
  }

  elements.toggleGraphButton.textContent = state.graphVisible ? "Скрыть граф" : "Показать граф";
}

function updateRoomStateClasses() {
  for (const room of state.rooms) {
    room.element.classList.toggle("is-selected", room.id === state.selectedRoomId);
  }
}

function assignSelectedVertex(target) {
  const vertex = getSelectedVertex();
  if (!vertex) {
    setRouteStatus("Сначала выберите узел графа.", "warning");
    return;
  }

  assignVertex(target, vertex.id);
}

function assignVertex(target, vertexID) {
  const vertex = findVertex(vertexID);
  if (!vertex) {
    return;
  }

  if (target === "start") {
    state.routeStartVertexId = vertex.id;
  } else {
    state.routeEndVertexId = vertex.id;
  }

  renderGraph();
  renderLists();
  renderSelection();
  renderRoutePlanner();
}

async function buildRoute() {
  const startVertex = getRouteStartVertex();
  const endVertex = getRouteEndVertex();
  if (!startVertex || !endVertex) {
    setRouteStatus("Нужно выбрать стартовый и конечный узел графа.", "warning");
    return;
  }

  const apiBase = normalizeApiBase(elements.apiBaseInput.value);
  const response = await fetch(`${apiBase}/routes`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      fromVertexId: startVertex.id,
      toVertexId: endVertex.id,
      accessibleOnly: elements.accessibleOnlyInput.checked,
    }),
  });

  if (!response.ok) {
    const errorPayload = await readErrorPayload(response);
    state.route = null;
    clearRouteOverlay();
    renderRoutePlanner();
    setRouteStatus(errorPayload, "error");
    return;
  }

  state.route = await response.json();
  renderRouteOverlay(state.route);
  renderRoutePlanner();
  focusRoute(state.route);
  setRouteStatus("Маршрут построен.", "ok");
}

function swapRouteVertices() {
  const previousStart = state.routeStartVertexId;
  state.routeStartVertexId = state.routeEndVertexId;
  state.routeEndVertexId = previousStart;
  renderGraph();
  renderLists();
  renderRoutePlanner();
}

function clearRoute() {
  state.route = null;
  state.routeStartVertexId = "";
  state.routeEndVertexId = "";
  clearRouteOverlay();
  renderGraph();
  renderLists();
  renderRoutePlanner();
  setRouteStatus("Маршрут очищен.", "ok");
}

function renderRoutePlanner() {
  const startVertex = getRouteStartVertex();
  const endVertex = getRouteEndVertex();

  elements.routeStartLabel.textContent = startVertex ? vertexDisplayName(startVertex) : "Не выбран";
  elements.routeEndLabel.textContent = endVertex ? vertexDisplayName(endVertex) : "Не выбран";
  elements.routeBadge.textContent = state.route ? "готов" : "ожидание";

  if (state.route) {
    elements.routeMetrics.innerHTML = `
      <span class="metric-chip">${state.route.distanceMeters} м</span>
      <span class="metric-chip">${state.route.estimatedDurationMinutes} мин</span>
      <span class="metric-chip">${state.route.steps.length} шагов</span>
    `;
    renderRouteSteps(state.route.steps);
  } else {
    elements.routeMetrics.replaceChildren();
    elements.routeSteps.replaceChildren();
  }

  updateRoutePlannerButtons();
}

function renderRouteSteps(steps) {
  const fragment = document.createDocumentFragment();
  for (const step of steps) {
    const item = document.createElement("div");
    item.className = "route-step";
    item.innerHTML = `
      <span class="route-step-order">${step.order}</span>
      <div>
        <p class="route-step-text">${escapeHTML(step.instruction)}</p>
        <p class="route-step-meta">
          ${escapeHTML(step.vertexId || "узел")} ·
          ${escapeHTML(`${formatNumber(step.coordinates.x)}, ${formatNumber(step.coordinates.y)}`)}
        </p>
      </div>
    `;
    fragment.append(item);
  }

  elements.routeSteps.replaceChildren(fragment);
}

function updateRoutePlannerButtons() {
  const selectedVertex = Boolean(getSelectedVertex());
  const hasStart = Boolean(getRouteStartVertex());
  const hasEnd = Boolean(getRouteEndVertex());

  elements.setStartButton.disabled = !selectedVertex;
  elements.setEndButton.disabled = !selectedVertex;
  elements.buildRouteButton.disabled = !(hasStart && hasEnd);
  elements.swapRouteButton.disabled = !(hasStart || hasEnd);
  elements.clearRouteButton.disabled = !(hasStart || hasEnd || state.route);

  elements.selectionCard.querySelector("#selectionStartButton")?.toggleAttribute("disabled", !selectedVertex);
  elements.selectionCard.querySelector("#selectionEndButton")?.toggleAttribute("disabled", !selectedVertex);
}

function renderRouteOverlay(route) {
  if (!state.svgElement) {
    return;
  }

  const group = state.svgElement.querySelector(`#${ROUTE_GROUP_ID}`);
  if (!group) {
    return;
  }

  group.replaceChildren();
  if (!route?.steps?.length) {
    return;
  }

  const polyline = document.createElementNS("http://www.w3.org/2000/svg", "polyline");
  polyline.setAttribute("class", "map-route-line");
  polyline.setAttribute(
    "points",
    route.steps.map((step) => `${step.coordinates.x},${step.coordinates.y}`).join(" ")
  );
  group.append(polyline);

  route.steps.forEach((step, index) => {
    const circle = document.createElementNS("http://www.w3.org/2000/svg", "circle");
    const roleClass =
      index === 0 ? "is-start" : index === route.steps.length - 1 ? "is-end" : "is-middle";
    circle.setAttribute("class", `map-route-point ${roleClass}`);
    circle.setAttribute("cx", String(step.coordinates.x));
    circle.setAttribute("cy", String(step.coordinates.y));
    circle.setAttribute("r", index === 0 || index === route.steps.length - 1 ? "8" : "5");
    group.append(circle);
  });
}

function clearRouteOverlay() {
  state.svgElement?.querySelector(`#${ROUTE_GROUP_ID}`)?.replaceChildren();
}

function focusRoute(route) {
  if (!state.svgElement || !route?.steps?.length) {
    return;
  }

  const xs = route.steps.map((step) => Number(step.coordinates.x));
  const ys = route.steps.map((step) => Number(step.coordinates.y));
  const minX = Math.min(...xs);
  const maxX = Math.max(...xs);
  const minY = Math.min(...ys);
  const maxY = Math.max(...ys);

  const width = Math.max(maxX - minX, 30);
  const height = Math.max(maxY - minY, 30);
  const paddingX = Math.max(50, width * 0.18);
  const paddingY = Math.max(50, height * 0.22);

  state.svgElement.setAttribute(
    "viewBox",
    [minX - paddingX, minY - paddingY, width + paddingX * 2, height + paddingY * 2]
      .map((value) => formatNumber(value))
      .join(" ")
  );
}

function resetViewBox() {
  if (!state.svgElement || !state.originalViewBox) {
    return;
  }

  state.svgElement.setAttribute("viewBox", state.originalViewBox);
}

function setStatus(message, tone) {
  elements.apiStatus.textContent = message;
  elements.apiStatus.classList.remove("is-error", "is-ok", "is-warning");
  if (tone === "error") {
    elements.apiStatus.classList.add("is-error");
  }
  if (tone === "ok") {
    elements.apiStatus.classList.add("is-ok");
  }
  if (tone === "warning") {
    elements.apiStatus.classList.add("is-warning");
  }
}

function setRouteStatus(message, tone) {
  elements.routeStatus.textContent = message;
  elements.routeStatus.classList.remove("is-error", "is-ok", "is-warning");
  if (tone === "error") {
    elements.routeStatus.classList.add("is-error");
  }
  if (tone === "ok") {
    elements.routeStatus.classList.add("is-ok");
  }
  if (tone === "warning") {
    elements.routeStatus.classList.add("is-warning");
  }
}

async function readErrorPayload(response) {
  try {
    const payload = await response.json();
    return payload.message || payload.error || `Маршрут не построен (${response.status})`;
  } catch {
    return `Маршрут не построен (${response.status})`;
  }
}

function getSelectedRoom() {
  return state.rooms.find((room) => room.id === state.selectedRoomId) || null;
}

function getSelectedVertex() {
  return findVertex(state.selectedVertexId);
}

function getRouteStartVertex() {
  return findVertex(state.routeStartVertexId);
}

function getRouteEndVertex() {
  return findVertex(state.routeEndVertexId);
}

function findVertex(vertexID) {
  return state.graph?.vertices.find((vertex) => vertex.id === vertexID) || null;
}

function vertexDisplayName(vertex) {
  return vertex.placeId || vertex.id;
}

function vertexSearchText(vertex) {
  return [vertex.id, vertex.placeId, vertex.building, String(vertex.floor)].join(" ").toLowerCase();
}

function normalizeApiBase(rawValue) {
  return rawValue.trim().replace(/\/+$/, "") || "http://localhost:8000";
}

function humanizeId(id) {
  return id
    .replace(/[-_]+/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

function formatNumber(value) {
  return Number(value).toFixed(1).replace(/\.0$/, "");
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}
