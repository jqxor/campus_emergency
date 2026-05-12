import L from "leaflet";

function $id(id) {
  return document.getElementById(id);
}

const AUTH_TOKEN_KEY = "campus_auth_token";
const AUTH_USER_KEY = "campus_auth_user";
const LOCAL_AUTH_USERS_KEY = "campus_local_users";
const CAMPUS_GEO_CONFIG_KEY = "campus_geo_config_v1";
const AMAP_WEB_KEY = String((import.meta && import.meta.env && import.meta.env.VITE_AMAP_KEY) || "").trim();

const authState = {
  token: "",
  user: null,
};

const NCIST_CENTER = [39.9549172, 116.7983350];
const NCIST_OUTLINE = [
  [39.9624208, 116.7951659],
  [39.9624208, 116.8013227],
  [39.9513732, 116.8013227],
  [39.9513732, 116.7951659],
  [39.9562000, 116.7949000],
];

const NCIST_KEY_POINTS = [
  { name: "南门", point: [39.9515268, 116.7986691], color: "#f97316" },
  { name: "北门", point: [39.9572022, 116.7989527], color: "#2563eb" },
  { name: "疏散集合点", point: [39.9551037, 116.7964282], color: "#0ea5e9" },
  { name: "教学楼", point: [39.9562000, 116.7972000], color: "#16a34a" },
];

const CAMPUS_POINT_TYPE_STYLE = {
  assembly: { color: "#0ea5e9", label: "集合点" },
  evacuation: { color: "#f97316", label: "疏散点" },
  risk: { color: "#ef4444", label: "风险点" },
  gate: { color: "#2563eb", label: "校门" },
  building: { color: "#16a34a", label: "建筑" },
};

const mapState = {
  pathMap: null,
  evacMap: null,
  adminMap: null,
  pathLayer: null,
  pathObstacleLayer: null,
  evacLayer: null,
  adminCampusOverlayLayer: null,
  pathCampusOverlayLayer: null,
  evacCampusOverlayLayer: null,
  adminEditLayer: null,
  adminBoundaryDraft: [],
  adminEditMode: "none",
  adminMapEventsBound: false,
  pathObstacles: [],
  mapPinsLayer: null,
  pathPlayback: null,
  evacPlayback: null,
  userLocation: null,
  pathUserMarker: null,
  evacUserMarker: null,
  pathDestination: null,
  evacDestination: null,
  pathDestinationMarker: null,
  evacDestinationMarker: null,
  liveNavigation: {
    active: false,
    context: "",
    watchId: null,
    mode: "car",
    destination: null,
    startedAt: 0,
    lastPoint: null,
    lastTimestamp: 0,
    totalMeters: 0,
    latestSpeedMps: 0,
    latestDistanceM: 0,
    latestEtaS: 0,
    latestDurationS: 0,
    latestPlannedDistanceM: 0,
    latestPlannedDurationS: 0,
    nextReplanAt: 0,
    planning: false,
    lastLogAt: 0,
  },
};

function clonePoint(point) {
  return [Number(point[0]), Number(point[1])];
}

function getDefaultCampusGeoConfig() {
  const points = NCIST_KEY_POINTS.map((item, idx) => {
    let type = "building";
    if (item.name.includes("集合")) type = "assembly";
    else if (item.name.includes("门")) type = "gate";
    return {
      id: "default_" + idx,
      name: item.name,
      type,
      point: clonePoint(item.point),
    };
  });

  return {
    center: clonePoint(NCIST_CENTER),
    outline: NCIST_OUTLINE.map((p) => clonePoint(p)),
    points,
    updatedAt: new Date().toISOString(),
  };
}

function sanitizeCampusGeoConfig(raw) {
  const fallback = getDefaultCampusGeoConfig();
  if (!raw || typeof raw !== "object") return fallback;

  const outline = Array.isArray(raw.outline)
    ? raw.outline
      .map((p) => [Number(p && p[0]), Number(p && p[1])])
      .filter((p) => Number.isFinite(p[0]) && Number.isFinite(p[1]))
    : [];

  const points = Array.isArray(raw.points)
    ? raw.points
      .map((item, idx) => {
        const lat = Number(item && item.point && item.point[0]);
        const lng = Number(item && item.point && item.point[1]);
        if (!Number.isFinite(lat) || !Number.isFinite(lng)) return null;
        const type = String((item && item.type) || "building");
        return {
          id: String((item && item.id) || ("point_" + idx)),
          name: String((item && item.name) || CAMPUS_POINT_TYPE_STYLE[type]?.label || "点位"),
          type: CAMPUS_POINT_TYPE_STYLE[type] ? type : "building",
          point: [lat, lng],
        };
      })
      .filter(Boolean)
    : [];

  const center = Array.isArray(raw.center)
    ? [Number(raw.center[0]), Number(raw.center[1])]
    : null;
  const validCenter = center && Number.isFinite(center[0]) && Number.isFinite(center[1])
    ? center
    : (outline.length ? outline[0] : clonePoint(NCIST_CENTER));

  return {
    center: validCenter,
    outline: outline.length >= 3 ? outline : fallback.outline,
    points,
    updatedAt: new Date().toISOString(),
  };
}

function loadCampusGeoConfig() {
  const raw = localStorage.getItem(CAMPUS_GEO_CONFIG_KEY);
  const parsed = parseJSON(raw || "null", null);
  return sanitizeCampusGeoConfig(parsed);
}

function saveCampusGeoConfig(config) {
  const data = sanitizeCampusGeoConfig(config);
  data.updatedAt = new Date().toISOString();
  localStorage.setItem(CAMPUS_GEO_CONFIG_KEY, JSON.stringify(data));
  return data;
}

function samplePathPoints() {
  return [
    [39.9586000, 116.7961000],
    [39.9576000, 116.7973000],
    [39.9564000, 116.7985000],
    [39.9554000, 116.7998000],
    [39.9543000, 116.8008000],
  ];
}

function sampleEvacuationPoints() {
  return {
    route: [
      [39.9569000, 116.7975000],
      [39.9561000, 116.7983000],
      [39.9553000, 116.7994000],
      [39.9545000, 116.8004000],
    ],
    risk: [39.9563000, 116.7980000],
    assembly: [39.9542000, 116.8006000],
  };
}

function getCampusOverlayLayerKey(overlayKey) {
  if (overlayKey === "path") return "pathCampusOverlayLayer";
  if (overlayKey === "evac") return "evacCampusOverlayLayer";
  return "adminCampusOverlayLayer";
}

function clearCampusOverlayByKey(overlayKey) {
  const stateKey = getCampusOverlayLayerKey(overlayKey);
  clearLayer(mapState[stateKey]);
  mapState[stateKey] = null;
}

function renderCampusOverlay(map, overlayKey) {
  if (!map) return;
  clearCampusOverlayByKey(overlayKey);

  const cfg = loadCampusGeoConfig();
  const group = L.layerGroup();

  if (Array.isArray(cfg.outline) && cfg.outline.length >= 3) {
    L.polygon(cfg.outline, {
      color: "#22c55e",
      weight: 2,
      fillColor: "#22c55e",
      fillOpacity: 0.08,
    }).addTo(group).bindTooltip("校园范围", { direction: "top", sticky: true });
  }

  (cfg.points || []).forEach((item) => {
    const style = CAMPUS_POINT_TYPE_STYLE[item.type] || CAMPUS_POINT_TYPE_STYLE.building;
    L.circleMarker(item.point, {
      radius: 7,
      color: style.color,
      fillColor: style.color,
      fillOpacity: 0.95,
      weight: 2,
    }).addTo(group).bindTooltip(item.name + "（" + style.label + "）", { direction: "top", sticky: true });
  });

  group.addTo(map);
  const stateKey = getCampusOverlayLayerKey(overlayKey);
  mapState[stateKey] = group;
}

function refreshAllCampusOverlays() {
  if (mapState.pathMap) renderCampusOverlay(mapState.pathMap, "path");
  if (mapState.evacMap) renderCampusOverlay(mapState.evacMap, "evac");
  if (mapState.adminMap) renderCampusOverlay(mapState.adminMap, "admin");
}

function ensureSatelliteMap(containerId, overlayKey = "admin") {
  const el = $id(containerId);
  if (!el) return null;

  const cfg = loadCampusGeoConfig();

  const map = L.map(containerId, {
    zoomControl: true,
    preferCanvas: true,
    zoomAnimation: false,
    fadeAnimation: false,
    markerZoomAnimation: false,
  }).setView(cfg.center || NCIST_CENTER, 17);
  // 高德卫星底图 + 注记图层
  L.tileLayer("https://webst0{s}.is.autonavi.com/appmaptile?style=6&x={x}&y={y}&z={z}", {
    subdomains: [1, 2, 3, 4],
    attribution: "© AutoNavi",
    maxZoom: 19,
    updateWhenIdle: true,
    keepBuffer: 1,
    detectRetina: false,
  }).addTo(map);

  L.tileLayer("https://webst0{s}.is.autonavi.com/appmaptile?style=8&x={x}&y={y}&z={z}", {
    subdomains: [1, 2, 3, 4],
    attribution: "© AutoNavi",
    maxZoom: 19,
    updateWhenIdle: true,
    keepBuffer: 1,
    detectRetina: false,
    pane: "overlayPane",
    zIndex: 450,
  }).addTo(map);

  renderCampusOverlay(map, overlayKey);

  setTimeout(() => map.invalidateSize(), 80);
  return map;
}

function normalizeRoutePoints(payload) {
  if (payload && Array.isArray(payload.points) && payload.points.length >= 2) {
    const parsed = payload.points
      .map((p) => [Number(p.lat), Number(p.lng)])
      .filter((p) => Number.isFinite(p[0]) && Number.isFinite(p[1]));
    if (parsed.length >= 2) return parsed;
  }

  const startLat = Number(payload && payload.start_lat);
  const startLng = Number(payload && payload.start_lng);
  const endLat = Number(payload && payload.end_lat);
  const endLng = Number(payload && payload.end_lng);
  if ([startLat, startLng, endLat, endLng].every(Number.isFinite)) {
    const midLat = (startLat + endLat) / 2 + 0.0004;
    const midLng = (startLng + endLng) / 2 - 0.0003;
    return [
      [startLat, startLng],
      [midLat, midLng],
      [endLat, endLng],
    ];
  }

  return samplePathPoints();
}

function parseLatLngText(text) {
  const raw = String(text || "").trim();
  if (!raw) return null;
  const parts = raw.split(/[，,\s]+/).filter(Boolean);
  if (parts.length < 2) return null;
  const a = Number(parts[0]);
  const b = Number(parts[1]);
  if (!Number.isFinite(a) || !Number.isFinite(b)) return null;
  const lat = Math.abs(a) <= 90 ? a : b;
  const lng = Math.abs(a) <= 90 ? b : a;
  if (!Number.isFinite(lat) || !Number.isFinite(lng)) return null;
  return [lat, lng];
}

function parseRangeMeters(text) {
  const raw = String(text || "").trim();
  if (!raw) return 35;
  const m = Number(raw.replace(/[^\d.]/g, ""));
  if (!Number.isFinite(m) || m <= 0) return 35;
  return m;
}

function haversineMeters(a, b) {
  if (!a || !b) return 0;
  const toRad = (x) => (x * Math.PI) / 180;
  const dLat = toRad(b[0] - a[0]);
  const dLng = toRad(b[1] - a[1]);
  const s1 = Math.sin(dLat / 2);
  const s2 = Math.sin(dLng / 2);
  const aa = s1 * s1 + Math.cos(toRad(a[0])) * Math.cos(toRad(b[0])) * s2 * s2;
  return 2 * 6371000 * Math.asin(Math.min(1, Math.sqrt(aa)));
}

function polylineDistanceMeters(points) {
  if (!Array.isArray(points) || points.length < 2) return 0;
  let sum = 0;
  for (let i = 1; i < points.length; i += 1) {
    sum += haversineMeters(points[i - 1], points[i]);
  }
  return sum;
}

function roundMeters(v) {
  return Number((Number(v || 0)).toFixed(1));
}

function formatMetersText(v) {
  return roundMeters(v) + " 米";
}

function formatDurationText(seconds) {
  const sec = Number(seconds || 0);
  if (!Number.isFinite(sec) || sec <= 0) return "";
  const mins = Math.max(1, Math.round(sec / 60));
  if (mins < 60) return mins + " 分钟";
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return h + " 小时" + (m ? (m + " 分") : "");
}

function pickRouteLabelPoint(points, fallback) {
  if (Array.isArray(points) && points.length) {
    return points[Math.floor(points.length / 2)];
  }
  return fallback || null;
}

function buildAmapNavigationUrl(start, end, mode) {
  const navMode = mode === "walk" ? "walk" : "car";
  const qs = new URLSearchParams();
  qs.set("from", formatAmapLngLat(start) + ",起点");
  qs.set("to", formatAmapLngLat(end) + ",终点");
  qs.set("mode", navMode);
  qs.set("policy", "1");
  qs.set("src", "campus_path_system");
  qs.set("coordinate", "gaode");
  qs.set("callnative", "1");
  return "https://uri.amap.com/navigation?" + qs.toString();
}

function openAmapNavigation(start, end, options) {
  if (!start || !end) return "";
  const opts = options || {};
  const url = buildAmapNavigationUrl(start, end, opts.mode);
  const popup = window.open(url, "_blank", "noopener");
  const distanceText = Number.isFinite(Number(opts.distanceMeters))
    ? formatMetersText(opts.distanceMeters)
    : "未知";
  const durationText = formatDurationText(opts.durationSeconds);

  logLocal(opts.label || "高德端到端导航", {
    provider: "高德导航",
    mode: opts.mode === "walk" ? "walk" : "car",
    planned_distance_m: Number.isFinite(Number(opts.distanceMeters)) ? roundMeters(opts.distanceMeters) : null,
    planned_duration_s: Number.isFinite(Number(opts.durationSeconds)) ? Math.round(Number(opts.durationSeconds)) : null,
    summary: durationText
      ? ("规划距离 " + distanceText + "，预计 " + durationText)
      : ("规划距离 " + distanceText),
    amap_url: url,
    popup_opened: !!popup,
  });

  if (!popup) {
    logLocal(opts.label || "高德端到端导航", "浏览器拦截了新窗口，请允许弹窗后重试");
  }

  return url;
}

function formatSpeedText(mps) {
  const v = Number(mps || 0);
  if (!Number.isFinite(v) || v <= 0) return "0 km/h";
  return (v * 3.6).toFixed(1) + " km/h";
}

function formatEtaClock(seconds) {
  const sec = Number(seconds || 0);
  if (!Number.isFinite(sec) || sec <= 0) return "--";
  const t = new Date(Date.now() + sec * 1000);
  const hh = String(t.getHours()).padStart(2, "0");
  const mm = String(t.getMinutes()).padStart(2, "0");
  const ss = String(t.getSeconds()).padStart(2, "0");
  return hh + ":" + mm + ":" + ss;
}

function getContextMap(context) {
  return context === "evac" ? mapState.evacMap : mapState.pathMap;
}

function getContextLayerKey(context) {
  return context === "evac" ? "evacLayer" : "pathLayer";
}

function getContextPlaybackKey(context) {
  return context === "evac" ? "evacPlayback" : "pathPlayback";
}

function getContextDestinationKey(context) {
  return context === "evac" ? "evacDestination" : "pathDestination";
}

function getContextDestinationMarkerKey(context) {
  return context === "evac" ? "evacDestinationMarker" : "pathDestinationMarker";
}

function getContextDestinationInputId(context) {
  return context === "evac" ? "evacDestinationInput" : "pathDestinationInput";
}

function getContextModeInputId(context) {
  return context === "evac" ? "evacNavMode" : "pathNavMode";
}

function getContextInfoPanelId(context) {
  return context === "evac" ? "evacLiveNavInfo" : "pathLiveNavInfo";
}

function getContextLabel(context) {
  return context === "evac" ? "疏散端到端导航" : "端到端路径导航";
}

function getModeByContext(context) {
  const modeEl = $id(getContextModeInputId(context));
  const mode = modeEl ? String(modeEl.value || "").trim() : "";
  return mode === "walk" ? "walk" : "car";
}

function setLiveNavInfo(context, text) {
  const el = $id(getContextInfoPanelId(context));
  if (el) {
    el.textContent = text;
  }
}

function setDestinationForContext(context, point, reason) {
  if (!Array.isArray(point) || point.length < 2) return;
  const lat = Number(point[0]);
  const lng = Number(point[1]);
  if (!Number.isFinite(lat) || !Number.isFinite(lng)) return;
  const map = getContextMap(context);
  const key = getContextDestinationKey(context);
  const markerKey = getContextDestinationMarkerKey(context);
  const input = $id(getContextDestinationInputId(context));

  mapState[key] = [lat, lng];
  if (input) {
    input.value = lat.toFixed(6) + "," + lng.toFixed(6);
  }

  if (map) {
    clearLayer(mapState[markerKey]);
    mapState[markerKey] = L.circleMarker([lat, lng], {
      radius: 8,
      color: "#dc2626",
      fillColor: "#dc2626",
      fillOpacity: 0.95,
      weight: 2,
    }).addTo(map).bindTooltip("导航终点", { direction: "top", sticky: true });
  }

  logLocal(getContextLabel(context), {
    action: "终点已设置",
    reason: reason || "manual",
    destination: { lat: Number(lat.toFixed(6)), lng: Number(lng.toFixed(6)) },
  });
}

function resolveDestinationForContext(context) {
  const key = getContextDestinationKey(context);
  const existing = mapState[key];
  if (Array.isArray(existing) && Number.isFinite(existing[0]) && Number.isFinite(existing[1])) {
    return existing;
  }

  const input = $id(getContextDestinationInputId(context));
  const parsed = parseLatLngText(input && input.value);
  if (parsed) {
    setDestinationForContext(context, parsed, "input");
    return parsed;
  }
  return null;
}

function chooseDestinationOnMap(context) {
  const map = getContextMap(context);
  if (!map) {
    logLocal(getContextLabel(context), "地图尚未初始化，无法在地图上选终点");
    return;
  }

  setLiveNavInfo(context, "请在地图上单击选择导航终点...");
  map.once("click", (event) => {
    const point = [Number(event.latlng.lat), Number(event.latlng.lng)];
    setDestinationForContext(context, point, "map_click");
    setLiveNavInfo(context, "终点已选择，可点击“端到端导航”开始实时导航");
  });
}

function renderLiveRouteByContext(context, start, end, points, plannedDistance, plannedDuration) {
  const map = getContextMap(context);
  if (!map || !Array.isArray(points) || points.length < 2) return;

  const layerKey = getContextLayerKey(context);
  clearLayer(mapState[layerKey]);

  const color = context === "evac" ? "#f97316" : "#2563eb";
  const group = L.layerGroup();
  L.polyline(points, {
    color,
    weight: 5,
    opacity: 0.95,
    dashArray: "10 8",
  }).addTo(group).bindTooltip("实时导航路径");

  L.circleMarker(start, {
    radius: 7,
    color: "#16a34a",
    fillColor: "#16a34a",
    fillOpacity: 0.95,
  }).addTo(group).bindTooltip("当前位置");

  L.circleMarker(end, {
    radius: 7,
    color: "#dc2626",
    fillColor: "#dc2626",
    fillOpacity: 0.95,
  }).addTo(group).bindTooltip("导航终点");

  const d = Number.isFinite(plannedDistance) && plannedDistance > 0
    ? plannedDistance
    : polylineDistanceMeters(points);
  const durationText = formatDurationText(plannedDuration);
  const labelPoint = pickRouteLabelPoint(points, end);
  if (labelPoint) {
    const labelText = durationText
      ? ("规划距离 " + formatMetersText(d) + " · 预计 " + durationText)
      : ("规划距离 " + formatMetersText(d));
    L.circleMarker(labelPoint, {
      radius: 5,
      color,
      fillColor: color,
      fillOpacity: 0.95,
      weight: 2,
    }).addTo(group).bindTooltip(labelText, { permanent: true, direction: "top" });
  }

  group.addTo(map);
  mapState[layerKey] = group;
  map.fitBounds(L.latLngBounds(points.concat([start, end])), { padding: [30, 30] });
}

function stopLiveNavigation(reason) {
  const session = mapState.liveNavigation;
  if (session.watchId !== null && navigator.geolocation && navigator.geolocation.clearWatch) {
    navigator.geolocation.clearWatch(session.watchId);
  }

  if (session.active) {
    const elapsedS = session.startedAt ? Math.max(0, (Date.now() - session.startedAt) / 1000) : 0;
    const summary = {
      status: reason || "实时导航已停止",
      total_distance_m: roundMeters(session.totalMeters || 0),
      avg_speed_kmh: elapsedS > 0 ? Number(((session.totalMeters / elapsedS) * 3.6).toFixed(1)) : 0,
      elapsed_s: Math.round(elapsedS),
    };
    logLocal(getContextLabel(session.context || "path"), summary);
    setLiveNavInfo(session.context || "path", summary.status + "\n累计移动距离: " + summary.total_distance_m + " 米\n平均速度: " + summary.avg_speed_kmh + " km/h");
  }

  mapState.liveNavigation = {
    active: false,
    context: "",
    watchId: null,
    mode: "car",
    destination: null,
    startedAt: 0,
    lastPoint: null,
    lastTimestamp: 0,
    totalMeters: 0,
    latestSpeedMps: 0,
    latestDistanceM: 0,
    latestEtaS: 0,
    latestDurationS: 0,
    latestPlannedDistanceM: 0,
    latestPlannedDurationS: 0,
    nextReplanAt: 0,
    planning: false,
    lastLogAt: 0,
  };
}

async function refreshLiveRouteIfNeeded(session, currentPoint) {
  const now = Date.now();
  if (session.planning || now < session.nextReplanAt) return;
  session.planning = true;
  session.nextReplanAt = now + 9000;

  try {
    const planned = await getAmapPlannedRoute(currentPoint, session.destination, mapState.pathObstacles);
    session.latestPlannedDistanceM = Number(planned.distance || 0);
    session.latestPlannedDurationS = Number(planned.duration || 0);
    renderLiveRouteByContext(session.context, currentPoint, session.destination, planned.points, planned.distance, planned.duration);
  } catch (err) {
    const fallback = [currentPoint, session.destination];
    renderLiveRouteByContext(session.context, currentPoint, session.destination, fallback, 0, 0);
    logLocal(getContextLabel(session.context), (err && err.message) || "实时重规划失败，已回退直线引导");
  } finally {
    session.planning = false;
  }
}

async function handleLiveNavigationPosition(session, rawPoint, rawSpeed, accuracy, timestamp) {
  if (!session.active) return;

  let point = rawPoint;
  try {
    point = await convertGpsToGcj(rawPoint);
  } catch (_) {
    point = rawPoint;
  }

  mapState.userLocation = point;
  const map = getContextMap(session.context);
  const markerKey = session.context === "evac" ? "evacUserMarker" : "pathUserMarker";
  if (map) {
    const oldMarker = mapState[markerKey];
    clearLayer(oldMarker);
    mapState[markerKey] = L.circleMarker(point, {
      radius: 8,
      color: "#7c3aed",
      fillColor: "#7c3aed",
      fillOpacity: 0.95,
      weight: 2,
    }).addTo(map).bindTooltip("我的位置", { direction: "top", sticky: true });
    map.panTo(point, { animate: true, duration: 0.25 });
  }

  if (session.lastPoint) {
    session.totalMeters += haversineMeters(session.lastPoint, point);
  }

  let speedMps = Number(rawSpeed);
  if (!(Number.isFinite(speedMps) && speedMps >= 0) && session.lastPoint && session.lastTimestamp) {
    const dt = Math.max(0.3, (timestamp - session.lastTimestamp) / 1000);
    speedMps = haversineMeters(session.lastPoint, point) / dt;
  }
  if (!(Number.isFinite(speedMps) && speedMps >= 0)) speedMps = 0;

  session.lastPoint = point;
  session.lastTimestamp = timestamp;
  session.latestSpeedMps = speedMps;

  await refreshLiveRouteIfNeeded(session, point);

  const remaining = session.latestPlannedDistanceM > 0
    ? session.latestPlannedDistanceM
    : haversineMeters(point, session.destination);
  session.latestDistanceM = remaining;

  let etaS = 0;
  if (speedMps > 0.25) {
    etaS = remaining / speedMps;
  } else if (session.latestPlannedDurationS > 0) {
    etaS = session.latestPlannedDurationS;
  }
  session.latestEtaS = etaS;

  const infoText = [
    "状态: 导航中",
    "当前位置: " + point[0].toFixed(6) + ", " + point[1].toFixed(6),
    "终点: " + session.destination[0].toFixed(6) + ", " + session.destination[1].toFixed(6),
    "移动速度: " + formatSpeedText(speedMps),
    "累计移动距离: " + formatMetersText(session.totalMeters),
    "剩余距离: " + formatMetersText(remaining),
    "预计到达时间: " + (etaS > 0 ? formatEtaClock(etaS) : "--"),
    "预计剩余时长: " + (etaS > 0 ? formatDurationText(etaS) : "--"),
    "定位精度: " + (Number.isFinite(accuracy) ? (Math.round(accuracy) + " 米") : "未知"),
  ].join("\n");
  setLiveNavInfo(session.context, infoText);

  if (Date.now() - session.lastLogAt > 6000) {
    session.lastLogAt = Date.now();
    logLocal(getContextLabel(session.context), {
      speed_kmh: Number((speedMps * 3.6).toFixed(1)),
      moved_distance_m: roundMeters(session.totalMeters),
      remaining_distance_m: roundMeters(remaining),
      eta_time: etaS > 0 ? formatEtaClock(etaS) : "--",
      eta_duration: etaS > 0 ? formatDurationText(etaS) : "--",
      accuracy_m: Number.isFinite(accuracy) ? Math.round(accuracy) : null,
    });
  }

  if (remaining <= 12) {
    stopLiveNavigation("已到达终点");
  }
}

async function startLiveNavigation(context) {
  const map = getContextMap(context);
  if (!map) {
    logLocal(getContextLabel(context), "地图尚未初始化，无法开始导航");
    return;
  }

  const destination = resolveDestinationForContext(context);
  if (!destination) {
    setLiveNavInfo(context, "请先输入终点坐标，或点击“地图点击选终点”后再开始导航");
    chooseDestinationOnMap(context);
    return;
  }

  if (!(await requestGpsPermission(getContextLabel(context) + " GPS权限"))) {
    return;
  }

  stopPlayback(getContextPlaybackKey(context));
  stopLiveNavigation("已切换到新的实时导航会话");

  const startPoint = await getCurrentUserLocation();
  const mode = getModeByContext(context);

  mapState.liveNavigation = {
    active: true,
    context,
    watchId: null,
    mode,
    destination,
    startedAt: Date.now(),
    lastPoint: null,
    lastTimestamp: 0,
    totalMeters: 0,
    latestSpeedMps: 0,
    latestDistanceM: 0,
    latestEtaS: 0,
    latestDurationS: 0,
    latestPlannedDistanceM: 0,
    latestPlannedDurationS: 0,
    nextReplanAt: 0,
    planning: false,
    lastLogAt: 0,
  };

  setLiveNavInfo(context, "实时导航启动中...");
  openAmapNavigation(startPoint, destination, {
    label: getContextLabel(context),
    mode,
    distanceMeters: 0,
    durationSeconds: 0,
  });

  await handleLiveNavigationPosition(mapState.liveNavigation, startPoint, 0, NaN, Date.now());

  const watchId = navigator.geolocation.watchPosition(
    (pos) => {
      const session = mapState.liveNavigation;
      if (!session.active || session.context !== context) return;
      const lat = Number(pos.coords && pos.coords.latitude);
      const lng = Number(pos.coords && pos.coords.longitude);
      if (!Number.isFinite(lat) || !Number.isFinite(lng)) return;
      const rawPoint = [lat, lng];
      handleLiveNavigationPosition(
        session,
        rawPoint,
        Number(pos.coords && pos.coords.speed),
        Number(pos.coords && pos.coords.accuracy),
        Number(pos.timestamp || Date.now())
      );
    },
    (err) => {
      logLocal(getContextLabel(context), (err && err.message) || "实时定位失败");
      stopLiveNavigation("定位中断，实时导航已停止");
    },
    {
      enableHighAccuracy: true,
      timeout: 10000,
      maximumAge: 1000,
    }
  );

  mapState.liveNavigation.watchId = watchId;
}

function formatAmapLngLat(point) {
  return Number(point[1]).toFixed(6) + "," + Number(point[0]).toFixed(6);
}

function obstacleToPolygon(obstacle) {
  const lat = Number(obstacle && obstacle.point && obstacle.point[0]);
  const lng = Number(obstacle && obstacle.point && obstacle.point[1]);
  const radius = Number(obstacle && obstacle.radius);
  if (!Number.isFinite(lat) || !Number.isFinite(lng) || !Number.isFinite(radius) || radius <= 0) {
    return "";
  }
  const dLat = radius / 111320;
  const dLng = radius / (111320 * Math.max(0.2, Math.cos((lat * Math.PI) / 180)));
  const p1 = (lng - dLng).toFixed(6) + "," + (lat - dLat).toFixed(6);
  const p2 = (lng + dLng).toFixed(6) + "," + (lat - dLat).toFixed(6);
  const p3 = (lng + dLng).toFixed(6) + "," + (lat + dLat).toFixed(6);
  const p4 = (lng - dLng).toFixed(6) + "," + (lat + dLat).toFixed(6);
  return [p1, p2, p3, p4, p1].join(";");
}

function buildAvoidPolygons(obstacles) {
  const parts = (obstacles || []).map((o) => obstacleToPolygon(o)).filter(Boolean);
  return parts.join("|");
}

function parseAmapPolyline(polylineText) {
  const text = String(polylineText || "").trim();
  if (!text) return [];
  const rawPoints = text.split(";");
  const result = [];
  let lastKey = "";
  rawPoints.forEach((item) => {
    const pair = item.split(",");
    if (pair.length < 2) return;
    const lng = Number(pair[0]);
    const lat = Number(pair[1]);
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) return;
    const key = lat.toFixed(6) + "," + lng.toFixed(6);
    if (key === lastKey) return;
    lastKey = key;
    result.push([lat, lng]);
  });
  return result;
}

function pickPayloadPoint(payload, prefix) {
  const lat = Number(payload && (payload[prefix + "_lat"] ?? payload[prefix + "Lat"]));
  const lng = Number(payload && (payload[prefix + "_lng"] ?? payload[prefix + "Lng"]));
  if (Number.isFinite(lat) && Number.isFinite(lng)) {
    return [lat, lng];
  }

  const obj = payload && payload[prefix];
  if (obj && typeof obj === "object") {
    const olat = Number(obj.lat);
    const olng = Number(obj.lng);
    if (Number.isFinite(olat) && Number.isFinite(olng)) {
      return [olat, olng];
    }
  }

  const text = typeof obj === "string" ? obj : payload && payload[prefix + "_point"];
  const parsed = parseLatLngText(text);
  if (parsed) return parsed;

  const maybeName = String(text || "").trim();
  if (!maybeName) return null;
  const hit = NCIST_KEY_POINTS.find((x) => x.name === maybeName);
  return hit ? hit.point : null;
}

function getStartEndPoints(payload) {
  let start = pickPayloadPoint(payload, "start");
  let end = pickPayloadPoint(payload, "end");
  if (!start || !end) {
    const demo = samplePathPoints();
    if (!start) start = demo[0];
    if (!end) end = demo[demo.length - 1];
  }
  return { start, end };
}

async function getAmapPlannedRoute(start, end, obstacles) {
  if (!AMAP_WEB_KEY) {
    throw new Error("缺少 VITE_AMAP_KEY，无法调用高德路径规划");
  }
  const avoidPolygons = buildAvoidPolygons(obstacles || []);
  const qs = new URLSearchParams();
  qs.set("key", AMAP_WEB_KEY);
  qs.set("origin", formatAmapLngLat(start));
  qs.set("destination", formatAmapLngLat(end));
  qs.set("extensions", "base");
  qs.set("strategy", "0");
  if (avoidPolygons) {
    qs.set("avoidpolygons", avoidPolygons);
  }

  const resp = await fetch("https://restapi.amap.com/v3/direction/driving?" + qs.toString(), { method: "GET" });
  if (!resp.ok) {
    throw new Error("高德路径规划请求失败");
  }
  const data = await resp.json();
  if (!data || data.status !== "1" || !data.route || !Array.isArray(data.route.paths) || !data.route.paths.length) {
    throw new Error((data && data.info) || "高德路径规划结果为空");
  }
  const best = data.route.paths[0];
  const points = [];
  (best.steps || []).forEach((step) => {
    points.push(...parseAmapPolyline(step && step.polyline));
  });
  if (points.length < 2 && best.polyline) {
    points.push(...parseAmapPolyline(best.polyline));
  }
  if (points.length < 2) {
    throw new Error("高德返回路径点不足");
  }
  return {
    points,
    distance: Number(best.distance || 0),
    duration: Number(best.duration || 0),
  };
}

function clearLayer(layerRef) {
  if (layerRef && typeof layerRef.remove === "function") {
    layerRef.remove();
  }
}

function ensureMapPinsLayer(map) {
  if (!map) return null;
  if (!mapState.mapPinsLayer) {
    mapState.mapPinsLayer = L.layerGroup().addTo(map);
  }
  return mapState.mapPinsLayer;
}

function stopPlayback(playbackKey) {
  const state = mapState[playbackKey];
  if (!state) return;
  if (state.timer) {
    clearInterval(state.timer);
  }
  mapState[playbackKey] = {
    timer: null,
    marker: state.marker || null,
    idx: state.idx || 0,
    points: state.points || [],
  };
}

function startPlayback(map, playbackKey, points, label, color) {
  if (!map || !Array.isArray(points) || points.length < 2) return;
  stopPlayback(playbackKey);

  const state = mapState[playbackKey] || {};
  const marker = state.marker || L.circleMarker(points[0], {
    radius: 8,
    color,
    fillColor: color,
    fillOpacity: 0.95,
    weight: 2,
  }).addTo(map).bindTooltip("导航位置", { direction: "top", sticky: true });

  marker.setLatLng(points[0]);
  mapState[playbackKey] = {
    timer: null,
    marker,
    idx: 0,
    points,
  };

  const total = points.length;
  const timer = setInterval(() => {
    const cur = mapState[playbackKey];
    if (!cur) return;
    cur.idx += 1;
    if (cur.idx >= total) {
      clearInterval(cur.timer);
      cur.timer = null;
      logLocal(label, "端到端导航完成");
      return;
    }
    const p = points[cur.idx];
    cur.marker.setLatLng(p);
    map.panTo(p, { animate: true, duration: 0.3 });
  }, 380);

  mapState[playbackKey].timer = timer;
  logLocal(label, "端到端导航已开始（地图回放）");
}

function fitMapToCampus(map, label) {
  if (!map) return;
  const cfg = loadCampusGeoConfig();
  const outline = Array.isArray(cfg.outline) && cfg.outline.length >= 3 ? cfg.outline : NCIST_OUTLINE;
  map.fitBounds(L.latLngBounds(outline), { padding: [30, 30] });
  logLocal(label, "已回到校园范围");
}

function addCenterPin(map, label) {
  if (!map) return;
  const layer = ensureMapPinsLayer(map);
  if (!layer) return;
  const center = map.getCenter();
  L.circleMarker([center.lat, center.lng], {
    radius: 6,
    color: "#0ea5e9",
    fillColor: "#0ea5e9",
    fillOpacity: 0.95,
    weight: 2,
  }).addTo(layer).bindTooltip("临时点", { direction: "top", sticky: true });
  logLocal(label, {
    lat: Number(center.lat.toFixed(6)),
    lng: Number(center.lng.toFixed(6)),
  });
}

async function measureToUser(map, label) {
  if (!map) return;
  try {
    if (!mapState.userLocation) {
      await getCurrentUserLocation();
    }
    const center = map.getCenter();
    const centerPoint = [center.lat, center.lng];
    const dist = haversineMeters(centerPoint, mapState.userLocation);
    logLocal(label, {
      center: {
        lat: Number(center.lat.toFixed(6)),
        lng: Number(center.lng.toFixed(6)),
      },
      my_location: {
        lat: Number(mapState.userLocation[0].toFixed(6)),
        lng: Number(mapState.userLocation[1].toFixed(6)),
      },
      distance_m: roundMeters(dist),
    });
  } catch (err) {
    logLocal(label, (err && err.message) || "测距失败");
  }
}

async function convertGpsToGcj(point) {
  const lat = Number(point && point[0]);
  const lng = Number(point && point[1]);
  if (!AMAP_WEB_KEY || !Number.isFinite(lat) || !Number.isFinite(lng)) {
    return [lat, lng];
  }

  const url = "https://restapi.amap.com/v3/assistant/coordinate/convert"
    + "?key=" + encodeURIComponent(AMAP_WEB_KEY)
    + "&locations=" + encodeURIComponent(lng + "," + lat)
    + "&coordsys=gps";

  const resp = await fetch(url, { method: "GET" });
  if (!resp.ok) {
    throw new Error("高德坐标转换请求失败");
  }
  const data = await resp.json();
  if (!data || data.status !== "1" || !data.locations) {
    throw new Error("高德坐标转换结果无效");
  }

  const parts = String(data.locations).split(",");
  const fixedLng = Number(parts[0]);
  const fixedLat = Number(parts[1]);
  if (!Number.isFinite(fixedLat) || !Number.isFinite(fixedLng)) {
    throw new Error("高德坐标解析失败");
  }
  return [fixedLat, fixedLng];
}

async function requestGpsPermission(label) {
  if (!navigator.geolocation) {
    logLocal(label, "当前浏览器不支持定位");
    return false;
  }

  if (!window.isSecureContext) {
    logLocal(label, "当前页面不是安全上下文（HTTPS/localhost），浏览器会阻止定位");
    return false;
  }

  try {
    if (navigator.permissions && navigator.permissions.query) {
      const status = await navigator.permissions.query({ name: "geolocation" });
      if (status.state === "granted") {
        logLocal(label, "定位权限已授权");
        return true;
      }
      if (status.state === "denied") {
        logLocal(label, "定位权限已被拒绝，请在浏览器站点设置中开启");
        return false;
      }
    }
  } catch (_) {
    // 忽略权限查询异常，继续走实际定位触发授权弹窗
  }

  try {
    await getCurrentUserLocation();
    logLocal(label, "已申请并获取GPS定位权限");
    return true;
  } catch (err) {
    logLocal(label, (err && err.message) || "定位权限申请失败");
    return false;
  }
}

function formatGeoError(err) {
  if (!err) return "定位失败";
  const code = Number(err.code);
  if (code === 1) {
    return "定位权限被拒绝，请在浏览器地址栏站点设置中允许位置权限";
  }
  if (code === 2) {
    return "当前位置不可用，请检查手机/系统定位服务是否开启";
  }
  if (code === 3) {
    return "定位超时，请在空旷区域重试，或先点击“申请GPS权限”";
  }
  const msg = String(err.message || "").trim();
  if (msg) {
    if (/Only secure origins|secure context|insecure/i.test(msg)) {
      return "浏览器仅允许在 HTTPS 或 localhost 下定位；请改为 HTTPS 访问";
    }
    return msg;
  }
  return "定位失败";
}

function getCurrentUserLocation() {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error("当前浏览器不支持定位"));
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const lat = Number(pos.coords && pos.coords.latitude);
        const lng = Number(pos.coords && pos.coords.longitude);
        if (!Number.isFinite(lat) || !Number.isFinite(lng)) {
          reject(new Error("定位结果无效"));
          return;
        }
        const rawPoint = [lat, lng];
        convertGpsToGcj(rawPoint)
          .then((fixedPoint) => {
            mapState.userLocation = fixedPoint;
            resolve(mapState.userLocation);
          })
          .catch(() => {
            mapState.userLocation = rawPoint;
            resolve(mapState.userLocation);
          });
      },
      (err) => {
        reject(new Error(formatGeoError(err)));
      },
      {
        enableHighAccuracy: true,
        timeout: 12000,
        maximumAge: 60000,
      }
    );
  });
}

function moveMapToUserLocation(map, markerKey, label) {
  const point = mapState.userLocation;
  if (!map || !point) return;
  const oldMarker = mapState[markerKey];
  clearLayer(oldMarker);
  mapState[markerKey] = L.circleMarker(point, {
    radius: 8,
    color: "#7c3aed",
    fillColor: "#7c3aed",
    fillOpacity: 0.95,
    weight: 2,
  }).addTo(map).bindTooltip("我的位置", { direction: "top", sticky: true });
  map.flyTo(point, Math.max(map.getZoom(), 18), { duration: 0.8 });
  logLocal(label, { lat: Number(point[0].toFixed(6)), lng: Number(point[1].toFixed(6)) });
}

async function locateAndCenter(map, markerKey, label) {
  if (!map) {
    logLocal(label, "地图尚未初始化，请先进入对应地图页面后再定位");
    return;
  }
  try {
    await getCurrentUserLocation();
    moveMapToUserLocation(map, markerKey, label);
  } catch (err) {
    if (Array.isArray(mapState.userLocation) && Number.isFinite(mapState.userLocation[0]) && Number.isFinite(mapState.userLocation[1])) {
      moveMapToUserLocation(map, markerKey, label + "（使用最近一次位置）");
      logLocal(label, "实时定位失败，已回退到最近一次定位结果");
      return;
    }
    logLocal(label, (err && err.message) || "定位失败");
  }
}

async function resetToUserLocation(map, markerKey, label) {
  if (!map) return;
  if (!mapState.userLocation) {
    await locateAndCenter(map, markerKey, label);
    return;
  }
  moveMapToUserLocation(map, markerKey, label);
}

function bindMapSidebarToggle(toggleBtnId, shellId) {
  const btn = $id(toggleBtnId);
  const shell = $id(shellId);
  if (!btn || !shell) return;

  const syncLabel = () => {
    const collapsed = shell.classList.contains("sidebar-collapsed");
    btn.textContent = collapsed ? "展开侧栏" : "收起侧栏";
  };

  syncLabel();
  btn.addEventListener("click", () => {
    shell.classList.toggle("sidebar-collapsed");
    syncLabel();
    const map = shellId === "pathMapShell" ? mapState.pathMap : mapState.evacMap;
    if (map) {
      setTimeout(() => map.invalidateSize(), 120);
    }
  });
}

function renderPathOverlay(points, title = "路径") {
  if (!mapState.pathMap) return;
  clearLayer(mapState.pathLayer);

  const group = L.layerGroup();
  L.polyline(points, { color: "#1d4ed8", weight: 5, opacity: 0.9 }).addTo(group);
  L.circleMarker(points[0], {
    radius: 7,
    color: "#16a34a",
    fillColor: "#16a34a",
    fillOpacity: 0.9,
  }).addTo(group).bindTooltip("起点");
  L.circleMarker(points[points.length - 1], {
    radius: 7,
    color: "#dc2626",
    fillColor: "#dc2626",
    fillOpacity: 0.9,
  }).addTo(group).bindTooltip("终点");

  group.addTo(mapState.pathMap);
  mapState.pathLayer = group;
  mapState.pathMap.fitBounds(L.latLngBounds(points), { padding: [30, 30] });
  logLocal("卫星路径标注", { title, points });
}

function ensurePathObstacleLayer() {
  if (!mapState.pathMap) return null;
  if (!mapState.pathObstacleLayer) {
    mapState.pathObstacleLayer = L.layerGroup().addTo(mapState.pathMap);
  }
  return mapState.pathObstacleLayer;
}

function addPathObstacle(point, radiusMeters, type, sourceText) {
  if (!Array.isArray(point) || point.length < 2) return;
  const obstacle = {
    point,
    radius: radiusMeters,
    type: String(type || "roadblock"),
    source: String(sourceText || ""),
  };
  mapState.pathObstacles.push(obstacle);
  const layer = ensurePathObstacleLayer();
  if (layer) {
    L.circle(point, {
      radius: radiusMeters,
      color: "#ef4444",
      fillColor: "#ef4444",
      fillOpacity: 0.15,
      weight: 2,
    }).addTo(layer).bindTooltip("障碍物: " + obstacle.type + " (" + Math.round(radiusMeters) + "m)");
  }
}

function renderPlannedPath(start, end, plannedPoints, title, plannedDistanceMeters, plannedDurationSeconds) {
  if (!mapState.pathMap) return;
  clearLayer(mapState.pathLayer);

  const group = L.layerGroup();
  const straightDistance = haversineMeters(start, end);

  // 实线：起终点最短直线
  L.polyline([start, end], {
    color: "#2563eb",
    weight: 4,
    opacity: 0.9,
  }).addTo(group).bindTooltip("最短直线");

  // 虚线：基于高德道路网络规划路径（可避障）
  L.polyline(plannedPoints, {
    color: "#f97316",
    weight: 5,
    opacity: 0.95,
    dashArray: "10 8",
  }).addTo(group).bindTooltip("路径规划（道路网络）");

  L.circleMarker(start, {
    radius: 7,
    color: "#16a34a",
    fillColor: "#16a34a",
    fillOpacity: 0.95,
  }).addTo(group).bindTooltip("起点");

  L.circleMarker(end, {
    radius: 7,
    color: "#dc2626",
    fillColor: "#dc2626",
    fillOpacity: 0.95,
  }).addTo(group).bindTooltip("终点");

  group.addTo(mapState.pathMap);
  mapState.pathLayer = group;
  mapState.pathMap.fitBounds(L.latLngBounds(plannedPoints.concat([start, end])), { padding: [30, 30] });

  const plannedDistance = Number.isFinite(plannedDistanceMeters) && plannedDistanceMeters > 0
    ? plannedDistanceMeters
    : polylineDistanceMeters(plannedPoints);
  const plannedDuration = Number.isFinite(plannedDurationSeconds) && plannedDurationSeconds > 0
    ? plannedDurationSeconds
    : 0;
  const labelPoint = pickRouteLabelPoint(plannedPoints, end);
  const durationText = formatDurationText(plannedDuration);
  if (labelPoint) {
    const labelText = durationText
      ? ("规划距离 " + formatMetersText(plannedDistance) + " · 预计 " + durationText)
      : ("规划距离 " + formatMetersText(plannedDistance));
    L.circleMarker(labelPoint, {
      radius: 5,
      color: "#f97316",
      fillColor: "#f97316",
      fillOpacity: 0.95,
      weight: 2,
    }).addTo(group).bindTooltip(labelText, { permanent: true, direction: "top" });
  }

  logLocal("卫星路径标注", {
    title,
    provider: "高德路径规划（道路网络）",
    obstacle_count: mapState.pathObstacles.length,
    straight_distance_m: roundMeters(straightDistance),
    planned_distance_m: roundMeters(plannedDistance),
    planned_duration_s: plannedDuration ? Math.round(plannedDuration) : 0,
    note: "实线=最短直线，虚线=路径规划（避让障碍物）",
  });

  return {
    start,
    end,
    points: plannedPoints,
    straightDistance,
    plannedDistance,
    plannedDuration,
  };
}

async function renderPathByPlanning(payload, title) {
  const { start, end } = getStartEndPoints(payload || {});
  try {
    const planned = await getAmapPlannedRoute(start, end, mapState.pathObstacles);
    return renderPlannedPath(start, end, planned.points, title, planned.distance, planned.duration);
  } catch (err) {
    const fallback = normalizeRoutePoints({
      start_lat: start[0],
      start_lng: start[1],
      end_lat: end[0],
      end_lng: end[1],
    });
    const result = renderPlannedPath(start, end, fallback, title + "（降级模式）", 0, 0);
    logLocal("路径规划降级", (err && err.message) || "高德路径规划失败，已回退演示路径");
    return result;
  }
}

function renderEvacuationOverlay(config) {
  if (!mapState.evacMap) return;
  clearLayer(mapState.evacLayer);

  const group = L.layerGroup();
  L.polyline(config.route, { color: "#f97316", weight: 5, dashArray: "8 6" }).addTo(group);
  L.circleMarker(config.risk, {
    radius: 8,
    color: "#ef4444",
    fillColor: "#ef4444",
    fillOpacity: 0.95,
  }).addTo(group).bindTooltip("风险点");
  L.circleMarker(config.assembly, {
    radius: 8,
    color: "#0ea5e9",
    fillColor: "#0ea5e9",
    fillOpacity: 0.95,
  }).addTo(group).bindTooltip("集合点");

  group.addTo(mapState.evacMap);
  mapState.evacLayer = group;
  mapState.evacMap.fitBounds(L.latLngBounds(config.route), { padding: [30, 30] });
}

async function renderEvacuationE2E() {
  const sample = sampleEvacuationPoints();
  const risk = sample.risk;
  const assembly = sample.assembly;
  let route = sample.route;
  let plannedDistance = 0;
  let plannedDuration = 0;

  try {
    const planned = await getAmapPlannedRoute(risk, assembly, mapState.pathObstacles);
    route = planned.points;
    plannedDistance = Number(planned.distance || 0);
    plannedDuration = Number(planned.duration || 0);
  } catch (err) {
    logLocal("疏散端到端规划", (err && err.message) || "高德规划失败，已使用演示路线");
  }

  clearLayer(mapState.evacLayer);
  const group = L.layerGroup();
  L.polyline(route, {
    color: "#f97316",
    weight: 5,
    dashArray: "10 8",
    opacity: 0.95,
  }).addTo(group).bindTooltip("疏散规划路径");

  L.circleMarker(risk, {
    radius: 8,
    color: "#ef4444",
    fillColor: "#ef4444",
    fillOpacity: 0.95,
  }).addTo(group).bindTooltip("风险点");

  L.circleMarker(assembly, {
    radius: 8,
    color: "#0ea5e9",
    fillColor: "#0ea5e9",
    fillOpacity: 0.95,
  }).addTo(group).bindTooltip("集合点");

  const effectiveDistance = plannedDistance || polylineDistanceMeters(route);
  const labelPoint = pickRouteLabelPoint(route, assembly);
  const durationText = formatDurationText(plannedDuration);
  if (labelPoint) {
    const labelText = durationText
      ? ("疏散距离 " + formatMetersText(effectiveDistance) + " · 预计 " + durationText)
      : ("疏散距离 " + formatMetersText(effectiveDistance));
    L.circleMarker(labelPoint, {
      radius: 5,
      color: "#f97316",
      fillColor: "#f97316",
      fillOpacity: 0.95,
      weight: 2,
    }).addTo(group).bindTooltip(labelText, { permanent: true, direction: "top" });
  }

  group.addTo(mapState.evacMap);
  mapState.evacLayer = group;
  mapState.evacMap.fitBounds(L.latLngBounds(route), { padding: [30, 30] });

  logLocal("疏散端到端规划", {
    risk,
    assembly,
    obstacle_count: mapState.pathObstacles.length,
    planned_distance_m: roundMeters(effectiveDistance),
    planned_duration_s: plannedDuration ? Math.round(plannedDuration) : 0,
  });

  startPlayback(mapState.evacMap, "evacPlayback", route, "疏散端到端导航", "#f97316");
  openAmapNavigation(risk, assembly, {
    label: "疏散端到端导航",
    mode: "walk",
    distanceMeters: effectiveDistance,
    durationSeconds: plannedDuration,
  });
}

function initCampusMaps(scopeEl) {
  const hasPath = scopeEl
    ? !!(scopeEl.querySelector && scopeEl.querySelector("#pathMap"))
    : !!$id("pathMap");
  const hasEvac = scopeEl
    ? !!(scopeEl.querySelector && scopeEl.querySelector("#evacMap"))
    : !!$id("evacMap");
  const hasAdmin = scopeEl
    ? !!(scopeEl.querySelector && scopeEl.querySelector("#adminCampusMap"))
    : !!$id("adminCampusMap");

  if (hasPath && !mapState.pathMap && $id("pathMap")) {
    mapState.pathMap = ensureSatelliteMap("pathMap", "path");
  }
  if (hasEvac && !mapState.evacMap && $id("evacMap")) {
    mapState.evacMap = ensureSatelliteMap("evacMap", "evac");
  }
  if (hasAdmin && !mapState.adminMap && $id("adminCampusMap")) {
    mapState.adminMap = ensureSatelliteMap("adminCampusMap", "admin");
  }
}

function invalidateVisibleMaps(scopeEl) {
  const ids = [];
  if (scopeEl && scopeEl.querySelector("#pathMap")) ids.push("path");
  if (scopeEl && scopeEl.querySelector("#evacMap")) ids.push("evac");
  if (scopeEl && scopeEl.querySelector("#adminCampusMap")) ids.push("admin");

  if (ids.includes("path") && mapState.pathMap) {
    setTimeout(() => mapState.pathMap.invalidateSize(), 60);
  }
  if (ids.includes("evac") && mapState.evacMap) {
    setTimeout(() => mapState.evacMap.invalidateSize(), 60);
  }
  if (ids.includes("admin") && mapState.adminMap) {
    setTimeout(() => mapState.adminMap.invalidateSize(), 60);
  }
}

function setAuthLocked(locked) {
  const root = document.documentElement;
  if (locked) {
    root.setAttribute("data-auth", "locked");
  } else {
    root.setAttribute("data-auth", "ready");
  }
}

function pickEmailFromUsername(username) {
  const u = String(username || "").trim();
  if (!u) return "";
  if (u.includes("@")) return u;
  return u + "@campus.edu";
}

function setAuthError(message) {
  const el = $id("authError");
  if (el) el.textContent = String(message || "");
}

function clearAuthError() {
  setAuthError("");
}

function validateEmail(email) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(String(email || "").trim());
}

function validatePassword(password) {
  const s = String(password || "");
  return s.length >= 8 && /[a-zA-Z]/.test(s) && /\d/.test(s);
}

function getLocalUsers() {
  const raw = localStorage.getItem(LOCAL_AUTH_USERS_KEY);
  const list = parseJSON(raw || "[]", []);
  return Array.isArray(list) ? list : [];
}

function saveLocalUsers(users) {
  localStorage.setItem(LOCAL_AUTH_USERS_KEY, JSON.stringify(users || []));
}

function findLocalUser(account) {
  const q = String(account || "").trim().toLowerCase();
  if (!q) return null;
  return getLocalUsers().find((u) => {
    const name = String(u.username || "").toLowerCase();
    const email = String(u.email || "").toLowerCase();
    return q === name || q === email;
  }) || null;
}

function createLocalUserRecord(payload) {
  return {
    id: Date.now(),
    username: payload.username,
    email: payload.email,
    password: payload.password,
    role_name: payload.role_name,
    profile_name: payload.profile_name,
    student_id: payload.student_id,
  };
}

function setSidebarUser(user) {
  const nameEl = document.querySelector(".side-user-name");
  const subEl = document.querySelector(".side-user-sub");
  if (nameEl) nameEl.textContent = user && user.username ? user.username : "未登录";
  if (subEl) {
    const role = user && user.role_name ? user.role_name : "";
    subEl.textContent = role ? role + " · 在线" : "离线";
  }
}

function activateView(targetId) {
  const buttons = Array.from(document.querySelectorAll(".menu-btn"));
  buttons.forEach((x) => x.classList.remove("active"));
  const menuBtn = document.querySelector('.menu-btn[data-target="' + targetId + '"]');
  if (menuBtn) menuBtn.classList.add("active");

  document.querySelectorAll(".view").forEach((view) => view.classList.remove("active"));
  const view = $id(targetId);
  if (view) view.classList.add("active");
}

function getTargetFromHash() {
  const hash = String(window.location.hash || "").trim();
  if (!hash || hash === "#") return "";
  const target = hash.replace(/^#/, "");
  const view = $id(target);
  if (!view || !view.classList.contains("view")) return "";
  return target;
}

function navigateToView(targetId, options = {}) {
  const target = String(targetId || "").trim();
  if (!target) return;
  const view = $id(target);
  if (!view || !view.classList.contains("view")) return;

  activateView(target);
  initCampusMaps(view);
  invalidateVisibleMaps(view);

  const shouldUpdateHash = options.updateHash !== false;
  if (shouldUpdateHash && window.location.hash !== "#" + target) {
    window.location.hash = target;
  }
}

function applyRoleIsolation(user) {
  const role = (user && user.role_name ? String(user.role_name) : "").toLowerCase();

  // 最小隔离策略：
  // - admin: 全部
  // - teacher: 无系统管理
  // - student: 仅首页/路径/预案/用户中心
  let allowed = new Set(["viewLogin", "viewHome", "viewUser", "viewPath", "viewEmergency", "viewReport", "viewMonitor", "viewSystem"]);
  if (role === "teacher") {
    allowed = new Set(["viewHome", "viewUser", "viewPath", "viewEmergency", "viewReport", "viewMonitor"]);
  } else if (role === "student") {
    allowed = new Set(["viewHome", "viewUser", "viewPath", "viewEmergency"]);
  }

  document.querySelectorAll(".menu-btn[data-target]").forEach((btn) => {
    const target = btn.dataset.target;
    if (!target) return;
    // 登录页在已登录后隐藏
    if (target === "viewLogin") {
      btn.hidden = true;
      return;
    }
    btn.hidden = !allowed.has(target);
  });

  document.querySelectorAll(".quick[data-jump]").forEach((btn) => {
    const target = btn.dataset.jump;
    const ok = target && allowed.has(target);

    if ("disabled" in btn) {
      btn.disabled = !ok;
    }
    btn.classList.toggle("quick-disabled", !ok);
    btn.setAttribute("aria-disabled", ok ? "false" : "true");
    if (ok) {
      btn.removeAttribute("tabindex");
    } else {
      btn.setAttribute("tabindex", "-1");
    }
  });

  // 如果当前 view 不允许，则回到首页
  const activeView = document.querySelector(".view.active");
  if (activeView && activeView.id && !allowed.has(activeView.id)) {
    navigateToView("viewHome", { updateHash: true });
  }
}

async function fetchJSON({ method, url, body, headers, isFormData = false }) {
  try {
    const reqHeaders = {
      ...(headers || {}),
    };
    if (authState.token && !reqHeaders.Authorization) {
      reqHeaders.Authorization = "Bearer " + authState.token;
    }
    if (!isFormData && body !== undefined && body !== null && !reqHeaders["Content-Type"]) {
      reqHeaders["Content-Type"] = "application/json";
    }

    const res = await fetch(url, {
      method,
      headers: reqHeaders,
      body: body
        ? isFormData
          ? body
          : typeof body === "string"
            ? body
            : JSON.stringify(body)
        : undefined,
    });
    const text = await res.text();
    const json = parseJSON(text, text);
    return { ok: res.ok, status: res.status, json };
  } catch (err) {
    return {
      ok: false,
      status: 0,
      json: {
        error: "NETWORK_ERROR",
        message: err && err.message ? err.message : String(err),
      },
    };
  }
}

function applyAuthSuccess(data, fallbackUser) {
  authState.token = (data && data.token) || "local-demo-token";
  authState.user = (data && data.user) || fallbackUser || null;
  if (authState.token) {
    localStorage.setItem(AUTH_TOKEN_KEY, authState.token);
  }
  if (authState.user) {
    localStorage.setItem(AUTH_USER_KEY, JSON.stringify(authState.user));
  }
  setAuthLocked(false);
  setSidebarUser(authState.user);
  applyRoleIsolation(authState.user);
  navigateToView("viewHome", { updateHash: true });
  updateMetrics();
}

function pretty(obj) {
  try {
    return JSON.stringify(obj, null, 2);
  } catch (_) {
    return String(obj);
  }
}

function parseJSON(text, fallback = null) {
  try {
    return JSON.parse(text);
  } catch (_) {
    return fallback;
  }
}

function toUint(val) {
  const n = Number(val);
  if (!Number.isFinite(n) || n < 0) return 0;
  return Math.floor(n);
}

function fmtDate(d) {
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function getNavDateQuery() {
  const startEl = document.getElementById("navStartDate");
  const endEl = document.getElementById("navEndDate");
  const start = startEl && startEl.value ? String(startEl.value).trim() : "";
  const end = endEl && endEl.value ? String(endEl.value).trim() : "";

  const now = new Date();
  const endDate = end || fmtDate(now);
  const startDate = start || fmtDate(new Date(now.getTime() - 29 * 24 * 60 * 60 * 1000));
  return `start_date=${encodeURIComponent(startDate)}&end_date=${encodeURIComponent(endDate)}`;
}

function logLocal(title, payload) {
  $id("output").textContent = [
    "[本地操作] " + title,
    "时间: " + new Date().toLocaleString(),
    "",
    typeof payload === "string" ? payload : pretty(payload),
  ].join("\n");
}

async function send({ method, url, body, headers, isFormData = false }) {
  const output = $id("output");
  output.textContent = "请求中...\n" + method + " " + url;

  try {
    const reqHeaders = {
      ...(headers || {}),
    };
    if (authState.token && !reqHeaders.Authorization) {
      reqHeaders.Authorization = "Bearer " + authState.token;
    }
    if (!isFormData) {
      reqHeaders["Content-Type"] = "application/json";
    }

    const res = await fetch(url, {
      method,
      headers: reqHeaders,
      body: body
        ? isFormData
          ? body
          : typeof body === "string"
            ? body
            : JSON.stringify(body)
        : undefined,
    });

    const text = await res.text();
    const json = parseJSON(text, text);
    output.textContent = [
      method + " " + url,
      "status: " + res.status,
      "",
      pretty(json),
    ].join("\n");

    return json;
  } catch (err) {
    const msg = err && err.message ? err.message : String(err);
    output.textContent = [
      method + " " + url,
      "status: (network error)",
      "",
      "请求失败: " + msg,
    ].join("\n");
    return { error: true, message: msg };
  }
}

function wireMenu() {
  document.addEventListener("click", (evt) => {
    const targetElement = evt.target && evt.target.closest
      ? evt.target.closest(".menu-btn[data-target], .quick[data-jump], a[href^='#']")
      : null;
    if (!targetElement) return;

    let target = "";
    if (targetElement.dataset && targetElement.dataset.target) {
      target = String(targetElement.dataset.target || "").trim();
    } else if (targetElement.dataset && targetElement.dataset.jump) {
      target = String(targetElement.dataset.jump || "").trim();
    } else if (targetElement.tagName === "A") {
      const href = String(targetElement.getAttribute("href") || "").trim();
      if (href.startsWith("#")) {
        target = href.slice(1).trim();
      }
    }

    if (!target) return;
    if (targetElement.getAttribute("aria-disabled") === "true") {
      evt.preventDefault();
      return;
    }

    const view = $id(target);
    if (!view || !view.classList.contains("view")) return;

    evt.preventDefault();
    navigateToView(target, { updateHash: true });
  });

  window.addEventListener("hashchange", () => {
    const target = getTargetFromHash();
    if (target) {
      navigateToView(target, { updateHash: false });
    }
  });
}

function wireSubtabs() {
  document.querySelectorAll(".subtabs").forEach((group) => {
    const buttons = Array.from(group.querySelectorAll(".subtab-btn"));
    buttons.forEach((btn) => {
      btn.addEventListener("click", () => {
        const parent = group.parentElement;
        const target = btn.dataset.subtab;
        buttons.forEach((x) => x.classList.remove("active"));
        btn.classList.add("active");

        parent.querySelectorAll(":scope .subtab").forEach((panel) => {
          panel.classList.remove("active");
        });
        const panel = parent.querySelector("#" + target);
        if (panel) {
          panel.classList.add("active");
          initCampusMaps(panel);
          invalidateVisibleMaps(panel);
        }
      });
    });
  });
}

function wireAuthTabs() {
  const buttons = Array.from(document.querySelectorAll(".auth-tab-btn"));
  const panels = Array.from(document.querySelectorAll(".auth-tab-panel"));
  if (!buttons.length || !panels.length) return;

  buttons.forEach((btn) => {
    btn.addEventListener("click", () => {
      const target = btn.dataset.authTab;
      buttons.forEach((x) => x.classList.remove("active"));
      panels.forEach((x) => x.classList.remove("active"));
      btn.classList.add("active");
      const panel = target ? $id(target) : null;
      if (panel) panel.classList.add("active");
      clearAuthError();
    });
  });
}

function bindLogin() {
  $id("btnLogin").addEventListener("click", async () => {
    const username = String($id("loginUser").value || "").trim();
    const password = String($id("loginPass").value || "").trim();
    const base = String($id("baseRole").value || "").trim();
    if (!username || !password) {
      setAuthError("请输入账号/邮箱与密码后再登录");
      return
    }
    clearAuthError();

    if (base) {
      const { ok, status, json } = await fetchJSON({
        method: "POST",
        url: base + "/api/auth/login",
        body: { username, password },
      });
      if (ok) {
        const data = json && json.data ? json.data : {};
        applyAuthSuccess(data, null);
        logLocal("账号登录", { username, role: authState.user && authState.user.role_name, result: "登录成功" });
        return;
      }

      // 明确鉴权失败时不回退本地
      if (status === 400 || status === 401 || status === 403) {
        setAuthError("账号或密码错误，请检查后重试");
        logLocal("账号登录失败", { status, error: (json && json.error) || json });
        return;
      }
    }

    const localUser = findLocalUser(username);
    if (localUser && String(localUser.password || "") === password) {
      applyAuthSuccess(null, {
        id: localUser.id,
        username: localUser.username,
        email: localUser.email,
        role_name: localUser.role_name || "student",
      });
      logLocal("账号登录", {
        username: localUser.username,
        result: "已通过本地注册账户登录（离线模式）",
      });
      return;
    }

    setAuthError("登录失败：账号不存在或密码不正确");
    logLocal("账号登录失败", "账号不存在或密码错误（本地/后端均未通过）");
  });

  $id("btnGetCode").addEventListener("click", () => {
    logLocal("验证码发送", { phone: $id("loginPhone").value, code: "123456" });
  });

  $id("btnCodeLogin").addEventListener("click", () => {
    logLocal("验证码登录", { phone: $id("loginPhone").value, status: "验证通过" });
  });

  $id("btnRegister").addEventListener("click", async () => {
    const profileName = String(($id("registerName") && $id("registerName").value) || "").trim();
    const studentId = String(($id("registerStudentId") && $id("registerStudentId").value) || "").trim();
    const username = String(($id("registerUser") && $id("registerUser").value) || "").trim();
    const password = String(($id("registerPass") && $id("registerPass").value) || "").trim();
    const confirmPassword = String(($id("registerConfirmPass") && $id("registerConfirmPass").value) || "").trim();
    const emailInput = String(($id("registerEmail") && $id("registerEmail").value) || "").trim();
    const roleName = String(($id("registerRole") && $id("registerRole").value) || "student").trim() || "student";
    const agree = !!($id("registerAgree") && $id("registerAgree").checked);
    const base = String($id("baseRole").value || "").trim();
    const email = emailInput || pickEmailFromUsername(username);

    if (!profileName) {
      setAuthError("请输入姓名");
      return
    }
    if (!studentId) {
      setAuthError("请输入学号");
      return
    }
    if (!/^\w{3,16}$/.test(username)) {
      setAuthError("用户名需为 3-16 位字母/数字/下划线");
      return
    }
    if (!validateEmail(email)) {
      setAuthError("请输入有效邮箱");
      return
    }
    if (!validatePassword(password)) {
      setAuthError("密码需至少8位，且包含字母和数字");
      return
    }
    if (password !== confirmPassword) {
      setAuthError("两次输入的密码不一致");
      return
    }
    if (!agree) {
      setAuthError("请先阅读并同意用户准则");
      return
    }

    clearAuthError();

    const exists = findLocalUser(username) || findLocalUser(email);
    if (exists) {
      setAuthError("该账号或邮箱已注册");
      return;
    }

    if (base) {
      const { ok, status, json } = await fetchJSON({
        method: "POST",
        url: base + "/api/auth/register",
        body: { username, password, email, role_name: roleName },
      });
      if (ok) {
        const data = json && json.data ? json.data : {};
        applyAuthSuccess(data, null);
        logLocal("用户注册", { username, role: authState.user && authState.user.role_name, result: "注册并登录成功" });
        return;
      }

      if (status === 400 || status === 409) {
        setAuthError("后端提示该账号已存在或参数无效");
        logLocal("用户注册失败", { status, error: (json && json.error) || json });
        return;
      }
    }

    const record = createLocalUserRecord({
      username,
      email,
      password,
      role_name: roleName,
      profile_name: profileName,
      student_id: studentId,
    });
    const users = getLocalUsers();
    users.push(record);
    saveLocalUsers(users);

    applyAuthSuccess(null, {
      id: record.id,
      username: record.username,
      email: record.email,
      role_name: record.role_name,
    });
    logLocal("用户注册", {
      username: record.username,
      role: record.role_name,
      result: "已完成本地注册并登录（离线模式）",
    });
  });
}

async function initAuth() {
  // 去登录化模式：默认使用本地管理员身份进入系统。
  authState.token = "local-demo-token";
  authState.user = {
    id: 1,
    username: "demo_admin",
    email: "demo_admin@campus.edu",
    role_name: "admin",
  };
  localStorage.setItem(AUTH_TOKEN_KEY, authState.token);
  localStorage.setItem(AUTH_USER_KEY, JSON.stringify(authState.user));

  setAuthLocked(false);
  setSidebarUser(authState.user);
  applyRoleIsolation(authState.user);
  const hashTarget = getTargetFromHash();
  navigateToView(hashTarget || "viewHome", { updateHash: false });
}

function bindPathModule() {
  $id("btnSaveAlgo").addEventListener("click", () => {
    logLocal("算法参数保存", {
      algorithm: $id("algoType").value,
      distance: $id("weightDistance").value,
      time: $id("weightTime").value,
      safety: $id("weightSafe").value,
    });
  });

  $id("btnImportPathCsv").addEventListener("click", () => {
    const file = $id("pathCsvFile").files[0];
    logLocal("路径CSV导入", file ? { filename: file.name } : "未选择文件");
  });

  $id("btnExportPathCsv").addEventListener("click", () => {
    logLocal("路径CSV导出", "已生成 path-settings.csv");
  });

  $id("btnPathSimulation").addEventListener("click", () => {
    logLocal("路径测试模拟", "Dijkstra: 72ms, A*: 49ms, 推荐 A*。");
  });

  $id("btnNavCalc").addEventListener("click", async () => {
    const payload = parseJSON($id("navCalcPayload").value, {});
    const endpoints = getStartEndPoints(payload);
    if (endpoints && endpoints.end) {
      setDestinationForContext("path", endpoints.end, "payload_end");
    }
    await send({
      method: "POST",
      url: $id("baseNav").value + "/api/navigation/path/calculate",
      body: payload,
      headers: { "X-User-ID": $id("navUserId").value },
    });
    await renderPathByPlanning(payload, "计算路径");
    updateMetrics();
  });

  $id("btnNavStart").addEventListener("click", async () => {
    await send({
      method: "POST",
      url: $id("baseNav").value + "/api/navigation/path/" + $id("navPathId").value + "/start",
      headers: { "X-User-ID": $id("navUserId").value },
    });
  });

  $id("btnNavUpdate").addEventListener("click", async () => {
    await send({
      method: "POST",
      url: $id("baseNav").value + "/api/navigation/path/" + $id("navPathId").value + "/update",
      body: parseJSON($id("navCalcPayload").value, {}),
      headers: { "X-User-ID": $id("navUserId").value },
    });
  });

  $id("btnNavEnd").addEventListener("click", async () => {
    if (mapState.liveNavigation.active && mapState.liveNavigation.context === "path") {
      stopLiveNavigation("路径实时导航已结束");
    }
    await send({
      method: "POST",
      url: $id("baseNav").value + "/api/navigation/path/" + $id("navPathId").value + "/end",
      headers: { "X-User-ID": $id("navUserId").value },
    });
  });

  $id("btnWarningConfirm").addEventListener("click", async () => {
    await send({
      method: "POST",
      url: $id("baseNav").value + "/api/navigation/warning/" + $id("warningId").value + "/confirm",
      headers: { "X-User-ID": $id("navUserId").value },
    });
  });

  $id("btnWarningIgnore").addEventListener("click", async () => {
    await send({
      method: "POST",
      url: $id("baseNav").value + "/api/navigation/warning/" + $id("warningId").value + "/ignore",
      headers: { "X-User-ID": $id("navUserId").value },
    });
  });

  $id("btnNavHistoryExport").addEventListener("click", async () => {
    await send({
      method: "GET",
      url: $id("baseNav").value + "/api/navigation/history/export?" + getNavDateQuery(),
      headers: { "X-User-ID": $id("navUserId").value },
    });
  });

  $id("btnNavSummary").addEventListener("click", async () => {
    await send({
      method: "GET",
      url: $id("baseNav").value + "/api/navigation/summary?" + getNavDateQuery(),
      headers: { "X-User-ID": $id("navUserId").value },
    });
  });

  $id("btnPathAnalyze").addEventListener("click", () => {
    logLocal("路径效率分析", "报告已生成并展示柱状图。");
  });

  $id("btnPathAdvice").addEventListener("click", () => {
    logLocal("优化建议", [
      "建议在教学区B增加分流路径。",
      "建议高峰期临时开放辅路。",
    ]);
  });

  $id("btnPathReportExport").addEventListener("click", () => {
    logLocal("导出路径报告", "已导出 path-analysis.xlsx");
  });

  $id("btnAddObstacle").addEventListener("click", () => {
    const tbody = $id("obstacleTable");
    const locationText = $id("obLocation").value;
    const obstacleType = $id("obType").value;
    const rangeText = $id("obRange").value;
    const tr = document.createElement("tr");
    tr.innerHTML = "<td>" + locationText + "</td><td>" + obstacleType + "</td><td>已新增</td>";
    tbody.appendChild(tr);

    const point = parseLatLngText(locationText);
    const rangeMeters = parseRangeMeters(rangeText);
    if (point) {
      addPathObstacle(point, rangeMeters, obstacleType, locationText);
    }

    logLocal("新增障碍物", {
      type: obstacleType,
      location: locationText,
      range: rangeText,
      map_applied: !!point,
      tip: point ? "已加入地图避障" : "位置需输入坐标，例如 39.9562,116.7972",
    });
  });

  $id("btnImportObstacleCsv").addEventListener("click", () => {
    const file = $id("obCsvFile").files[0];
    logLocal("障碍物CSV导入", file ? file.name : "未选择文件");
  });

  const btnPathRenderMap = $id("btnPathRenderMap");
  if (btnPathRenderMap) {
    btnPathRenderMap.addEventListener("click", async () => {
      const destination = resolveDestinationForContext("path");
      if (!destination) {
        setLiveNavInfo("path", "请先在地图标注终点（点击“地图点击选终点”）");
        chooseDestinationOnMap("path");
        logLocal("卫星路径标注", "未设置终点，已进入地图点选终点模式");
        return;
      }

      let startPoint = mapState.userLocation;
      if (!startPoint) {
        try {
          startPoint = await getCurrentUserLocation();
        } catch (err) {
          setLiveNavInfo("path", "无法获取我的位置，请先点击“申请GPS权限”或“定位到我”");
          logLocal("卫星路径标注", (err && err.message) || "获取我的位置失败");
          return;
        }
      }

      if (mapState.pathMap) {
        moveMapToUserLocation(mapState.pathMap, "pathUserMarker", "手动标注路径起点");
      }

      const payload = {
        start_lat: startPoint[0],
        start_lng: startPoint[1],
        end_lat: destination[0],
        end_lng: destination[1],
      };

      await renderPathByPlanning(payload, "手动标注路径（我 -> 标注点）");
    });
  }

  const btnPathE2ENav = $id("btnPathE2ENav");
  if (btnPathE2ENav) {
    btnPathE2ENav.addEventListener("click", async () => {
      await startLiveNavigation("path");
    });
  }

  const btnPathPickDestination = $id("btnPathPickDestination");
  if (btnPathPickDestination) {
    btnPathPickDestination.addEventListener("click", () => {
      chooseDestinationOnMap("path");
    });
  }

  const pathDestinationInput = $id("pathDestinationInput");
  if (pathDestinationInput) {
    pathDestinationInput.addEventListener("change", () => {
      const parsed = parseLatLngText(pathDestinationInput.value);
      if (parsed) {
        setDestinationForContext("path", parsed, "input_change");
      }
    });
  }

  const btnPathLiveNavStop = $id("btnPathLiveNavStop");
  if (btnPathLiveNavStop) {
    btnPathLiveNavStop.addEventListener("click", () => {
      if (mapState.liveNavigation.active && mapState.liveNavigation.context === "path") {
        stopLiveNavigation("路径实时导航已手动停止");
      }
    });
  }

  const btnPathClearMap = $id("btnPathClearMap");
  if (btnPathClearMap) {
    btnPathClearMap.addEventListener("click", () => {
      stopPlayback("pathPlayback");
      if (mapState.liveNavigation.active && mapState.liveNavigation.context === "path") {
        stopLiveNavigation("路径实时导航因清空地图而停止");
      }
      clearLayer(mapState.pathLayer);
      mapState.pathLayer = null;
      logLocal("卫星路径标注", "已清空路径标注");
    });
  }

  const btnPathLocate = $id("btnPathLocate");
  if (btnPathLocate) {
    btnPathLocate.addEventListener("click", async () => {
      await locateAndCenter(mapState.pathMap, "pathUserMarker", "路径地图定位");
    });
  }

  const btnPathRequestGps = $id("btnPathRequestGps");
  if (btnPathRequestGps) {
    btnPathRequestGps.addEventListener("click", async () => {
      await requestGpsPermission("路径地图GPS权限");
    });
  }

  const btnPathResetToMe = $id("btnPathResetToMe");
  if (btnPathResetToMe) {
    btnPathResetToMe.addEventListener("click", async () => {
      await resetToUserLocation(mapState.pathMap, "pathUserMarker", "路径地图重置");
    });
  }

  const btnPathCenterPin = $id("btnPathCenterPin");
  if (btnPathCenterPin) {
    btnPathCenterPin.addEventListener("click", () => {
      addCenterPin(mapState.pathMap, "路径地图中心打点");
    });
  }

  const btnPathMeasureToMe = $id("btnPathMeasureToMe");
  if (btnPathMeasureToMe) {
    btnPathMeasureToMe.addEventListener("click", async () => {
      await measureToUser(mapState.pathMap, "路径地图测距到我");
    });
  }

  const btnPathFitCampus = $id("btnPathFitCampus");
  if (btnPathFitCampus) {
    btnPathFitCampus.addEventListener("click", () => {
      fitMapToCampus(mapState.pathMap, "路径地图回到校园");
    });
  }

  bindMapSidebarToggle("btnPathSidebarToggle", "pathMapShell");
}

function bindEmergencyModule() {
  const sample = sampleEvacuationPoints();
  setDestinationForContext("evac", sample.assembly, "default_assembly_point");

  $id("btnPlanCreate").addEventListener("click", async () => {
    await send({
      method: "POST",
      url: $id("basePlan").value + "/api/plans",
      body: parseJSON($id("planCreatePayload").value, {}),
    });
    updateMetrics();
  });

  $id("btnPlanUpdate").addEventListener("click", async () => {
    await send({
      method: "PUT",
      url: $id("basePlan").value + "/api/plans/" + $id("planId").value,
      body: parseJSON($id("planCreatePayload").value, {}),
    });
  });

  $id("btnPlanDelete").addEventListener("click", async () => {
    await send({
      method: "DELETE",
      url: $id("basePlan").value + "/api/plans/" + $id("planId").value,
    });
  });

  $id("btnPlanSearch").addEventListener("click", async () => {
    const p = new URLSearchParams();
    if ($id("planScenario").value) p.set("scenario_type", $id("planScenario").value);
    if ($id("planStatus").value) p.set("status", $id("planStatus").value);
    if ($id("planKeyword").value) p.set("keyword", $id("planKeyword").value);
    p.set("page", "1");
    p.set("page_size", "10");
    await send({ method: "GET", url: $id("basePlan").value + "/api/plans/search?" + p.toString() });
  });

  $id("btnPlanGet").addEventListener("click", async () => {
    await send({ method: "GET", url: $id("basePlan").value + "/api/plans/" + $id("planId").value });
  });

  $id("btnPlanUpdateStatus").addEventListener("click", async () => {
    await send({
      method: "PATCH",
      url: $id("basePlan").value + "/api/plans/" + $id("planId").value + "/status",
      body: parseJSON($id("planStatusBody").value, { status: "active" }),
    });
  });

  $id("btnPlanImport").addEventListener("click", async () => {
    const file = $id("planImportFile").files[0];
    if (!file) {
      logLocal("预案导入", "请先选择文件");
      return;
    }
    const fd = new FormData();
    fd.append("file", file);
    await send({
      method: "POST",
      url: $id("basePlan").value + "/api/plans/import",
      body: fd,
      isFormData: true,
    });
  });

  $id("btnPlanExport").addEventListener("click", async () => {
    const p = new URLSearchParams();
    if ($id("planScenario").value) p.set("scenario_type", $id("planScenario").value);
    if ($id("planStatus").value) p.set("status", $id("planStatus").value);
    await send({ method: "GET", url: $id("basePlan").value + "/api/plans/export?" + p.toString() });
  });

  $id("btnRunSimulation").addEventListener("click", () => {
    const progress = $id("simProgress");
    progress.value = 0;
    const timer = setInterval(() => {
      progress.value += 10;
      if (progress.value >= 100) {
        clearInterval(timer);
        logLocal("实时疏散模拟", {
          planId: $id("simPlanId").value,
          people: $id("simPeople").value,
          speed: $id("simSpeed").value,
          result: "模拟完成，拥堵点2处，建议调整B区出口",
        });
        renderEvacuationOverlay(sampleEvacuationPoints());
      }
    }, 120);
  });

  $id("btnSimReport").addEventListener("click", () => {
    logLocal("模拟报告导出", "已导出 evacuation-simulation.pdf");
  });

  $id("btnTriggerEvent").addEventListener("click", () => {
    const line = new Date().toLocaleTimeString() + " | 已触发" + $id("eventType").value + "响应，预案ID=" + $id("eventPlanId").value;
    $id("eventLog").textContent = line;
    logLocal("事件触发", line);
  });

  $id("btnSendNotice").addEventListener("click", () => {
    logLocal("通知发送", "已向师生与应急团队发送疏散路线通知。");
  });

  $id("btnSubmitFeedback").addEventListener("click", () => {
    logLocal("反馈提交", $id("eventFeedback").value);
  });

  $id("btnPlanOptimize").addEventListener("click", async () => {
    await send({ method: "POST", url: $id("basePlan").value + "/api/plans/" + $id("optPlanId").value + "/optimize" });
  });

  $id("btnApplyPath").addEventListener("click", () => {
    logLocal("应用优化路径", {
      planId: $id("optPlanId").value,
      distanceWeight: $id("optDist").value,
      safetyWeight: $id("optSafe").value,
      status: "已应用",
    });
  });

  const btnEvacRenderMap = $id("btnEvacRenderMap");
  if (btnEvacRenderMap) {
    btnEvacRenderMap.addEventListener("click", () => {
      renderEvacuationOverlay(sampleEvacuationPoints());
      logLocal("疏散路线标注", "已在卫星图标注疏散路线、风险点与集合点");
    });
  }

  const btnEvacE2ENav = $id("btnEvacE2ENav");
  if (btnEvacE2ENav) {
    btnEvacE2ENav.addEventListener("click", async () => {
      await startLiveNavigation("evac");
    });
  }

  const btnEvacPickDestination = $id("btnEvacPickDestination");
  if (btnEvacPickDestination) {
    btnEvacPickDestination.addEventListener("click", () => {
      chooseDestinationOnMap("evac");
    });
  }

  const evacDestinationInput = $id("evacDestinationInput");
  if (evacDestinationInput) {
    evacDestinationInput.addEventListener("change", () => {
      const parsed = parseLatLngText(evacDestinationInput.value);
      if (parsed) {
        setDestinationForContext("evac", parsed, "input_change");
      }
    });
  }

  const btnEvacLiveNavStop = $id("btnEvacLiveNavStop");
  if (btnEvacLiveNavStop) {
    btnEvacLiveNavStop.addEventListener("click", () => {
      if (mapState.liveNavigation.active && mapState.liveNavigation.context === "evac") {
        stopLiveNavigation("疏散实时导航已手动停止");
      }
    });
  }

  const btnEvacClearMap = $id("btnEvacClearMap");
  if (btnEvacClearMap) {
    btnEvacClearMap.addEventListener("click", () => {
      stopPlayback("evacPlayback");
      if (mapState.liveNavigation.active && mapState.liveNavigation.context === "evac") {
        stopLiveNavigation("疏散实时导航因清空地图而停止");
      }
      clearLayer(mapState.evacLayer);
      mapState.evacLayer = null;
      logLocal("疏散路线标注", "已清空疏散标注");
    });
  }

  const btnEvacLocate = $id("btnEvacLocate");
  if (btnEvacLocate) {
    btnEvacLocate.addEventListener("click", async () => {
      await locateAndCenter(mapState.evacMap, "evacUserMarker", "疏散地图定位");
    });
  }

  const btnEvacRequestGps = $id("btnEvacRequestGps");
  if (btnEvacRequestGps) {
    btnEvacRequestGps.addEventListener("click", async () => {
      await requestGpsPermission("疏散地图GPS权限");
    });
  }

  const btnEvacResetToMe = $id("btnEvacResetToMe");
  if (btnEvacResetToMe) {
    btnEvacResetToMe.addEventListener("click", async () => {
      await resetToUserLocation(mapState.evacMap, "evacUserMarker", "疏散地图重置");
    });
  }

  const btnEvacCenterPin = $id("btnEvacCenterPin");
  if (btnEvacCenterPin) {
    btnEvacCenterPin.addEventListener("click", () => {
      addCenterPin(mapState.evacMap, "疏散地图中心打点");
    });
  }

  const btnEvacMeasureToMe = $id("btnEvacMeasureToMe");
  if (btnEvacMeasureToMe) {
    btnEvacMeasureToMe.addEventListener("click", async () => {
      await measureToUser(mapState.evacMap, "疏散地图测距到我");
    });
  }

  const btnEvacFitCampus = $id("btnEvacFitCampus");
  if (btnEvacFitCampus) {
    btnEvacFitCampus.addEventListener("click", () => {
      fitMapToCampus(mapState.evacMap, "疏散地图回到校园");
    });
  }

  bindMapSidebarToggle("btnEvacSidebarToggle", "evacMapShell");
}

function bindReportModule() {
  $id("btnRptGenerate").addEventListener("click", () => {
    logLocal("路径使用报表", {
      start: $id("rpStart").value,
      end: $id("rpEnd").value,
      type: $id("rpType").value,
      status: "已生成",
    });
  });

  $id("btnRptEvaluate").addEventListener("click", () => {
    logLocal("效率评估", "路径区域B效率指数: 81.2，建议新增东侧连通线。");
  });

  $id("btnRptExport").addEventListener("click", () => {
    logLocal("报表导出", "已导出 usage-report.pdf 与 usage-report.xlsx");
  });

  $id("btnEvReport").addEventListener("click", () => {
    logLocal("响应时间报告", { eventType: $id("evType").value, avg: "8.4分钟" });
  });

  $id("btnEvSuccess").addEventListener("click", () => {
    logLocal("成功率评估", { scene: $id("evScene").value, success: "95%" });
  });

  $id("btnEvStartSim").addEventListener("click", () => {
    logLocal("疏散模拟", "模拟执行中...完成，建议加强食堂区域分流。 ");
  });

  $id("btnEvExport").addEventListener("click", () => {
    logLocal("疏散报表导出", "已导出 evacuation-efficiency.csv / .pdf");
  });
}

function bindMonitorModule() {
  $id("btnCamSelect").addEventListener("click", () => {
    logLocal("摄像头切换", {
      area: $id("camArea").value,
      cameraId: $id("camId").value,
      mode: $id("camMode").value,
    });
  });

  $id("btnSnapshot").addEventListener("click", () => {
    logLocal("画面导出", "已保存 snapshot.png");
  });

  $id("btnTrackStart").addEventListener("click", () => {
    const msg = "路径追踪启动: " + $id("trackType").value + "，更新频率" + $id("trackFreq").value + "秒";
    $id("liveAlert").textContent = msg;
    logLocal("路径追踪", msg);
  });

  $id("btnRiskPredict").addEventListener("click", () => {
    logLocal("风险预测设置", {
      type: $id("riskType").value,
      threshold: $id("riskThreshold").value,
      cycle: $id("riskCycle").value,
    });
  });

  $id("btnRiskTest").addEventListener("click", () => {
    const tbody = $id("warnLog");
    const tr = document.createElement("tr");
    tr.innerHTML = "<td>" + new Date().toLocaleTimeString() + "</td><td>教学楼南门</td><td>高</td><td>待处理</td>";
    tbody.prepend(tr);
    logLocal("预警测试", "已触发高等级弹窗/声音/通知");
  });

  $id("btnWarnExport").addEventListener("click", () => {
    logLocal("导出预警日志", "已导出 warning-log.csv");
  });

  $id("btnMrGenerate").addEventListener("click", () => {
    logLocal("综合分析报告", {
      type: $id("mrType").value,
      range: $id("mrRange").value,
      status: "生成完成",
    });
  });

  $id("btnMrEdit").addEventListener("click", () => {
    logLocal("报告模板编辑", "已进入模板编辑模式");
  });

  $id("btnMrExport").addEventListener("click", () => {
    logLocal("报告导出", "已导出 monitor-report.pdf / .xlsx");
  });
}

function isAdminUser() {
  const role = String(authState.user && authState.user.role_name || "").toLowerCase();
  return role === "admin";
}

function ensureAdminEditLayer() {
  if (!mapState.adminMap) return null;
  if (!mapState.adminEditLayer) {
    mapState.adminEditLayer = L.layerGroup().addTo(mapState.adminMap);
  }
  return mapState.adminEditLayer;
}

function updateAdminGeoSummary() {
  const el = $id("adminGeoSummary");
  if (!el) return;
  const cfg = loadCampusGeoConfig();
  const countByType = {};
  (cfg.points || []).forEach((p) => {
    const t = String(p.type || "building");
    countByType[t] = (countByType[t] || 0) + 1;
  });

  const lines = [
    "更新时间: " + String(cfg.updatedAt || "-"),
    "范围点数: " + ((cfg.outline && cfg.outline.length) || 0),
    "点位总数: " + ((cfg.points && cfg.points.length) || 0),
    "集合点: " + (countByType.assembly || 0),
    "疏散点: " + (countByType.evacuation || 0),
    "风险点: " + (countByType.risk || 0),
  ];
  el.textContent = lines.join("\n");
}

function renderAdminBoundaryDraft() {
  const layer = ensureAdminEditLayer();
  if (!layer) return;
  layer.clearLayers();
  if (mapState.adminBoundaryDraft.length < 1) return;

  mapState.adminBoundaryDraft.forEach((p, idx) => {
    L.circleMarker(p, {
      radius: 5,
      color: "#2563eb",
      fillColor: "#2563eb",
      fillOpacity: 0.95,
      weight: 2,
    }).addTo(layer).bindTooltip("边界点 " + (idx + 1), { direction: "top" });
  });

  if (mapState.adminBoundaryDraft.length >= 2) {
    L.polyline(mapState.adminBoundaryDraft, {
      color: "#2563eb",
      weight: 3,
      dashArray: "8 6",
      opacity: 0.95,
    }).addTo(layer).bindTooltip("校园范围草稿");
  }
}

function setAdminEditMode(mode) {
  mapState.adminEditMode = mode;
  const modeText = mode === "boundary"
    ? "范围框选模式"
    : mode === "point"
      ? "点位新增模式"
      : "未编辑";
  logLocal("校园范围配置", "当前模式: " + modeText);
}

function ensureAdminMapEventsBound() {
  if (!mapState.adminMap || mapState.adminMapEventsBound) return;
  mapState.adminMap.on("click", (e) => {
    if (!isAdminUser()) {
      logLocal("校园范围配置", "仅 admin 可编辑校园范围");
      return;
    }

    const point = [Number(e.latlng.lat), Number(e.latlng.lng)];
    if (mapState.adminEditMode === "boundary") {
      mapState.adminBoundaryDraft.push(point);
      renderAdminBoundaryDraft();
      return;
    }

    if (mapState.adminEditMode === "point") {
      const cfg = loadCampusGeoConfig();
      const type = String(($id("adminPointType") && $id("adminPointType").value) || "assembly");
      const style = CAMPUS_POINT_TYPE_STYLE[type] || CAMPUS_POINT_TYPE_STYLE.assembly;
      const customName = String(($id("adminPointName") && $id("adminPointName").value) || "").trim();
      cfg.points.push({
        id: "pt_" + Date.now(),
        name: customName || (style.label + "-" + (cfg.points.length + 1)),
        type: style === CAMPUS_POINT_TYPE_STYLE[type] ? type : "assembly",
        point,
      });
      saveCampusGeoConfig(cfg);
      refreshAllCampusOverlays();
      updateAdminGeoSummary();
      logLocal("新增校园点位", cfg.points[cfg.points.length - 1]);
    }
  });
  mapState.adminMapEventsBound = true;
}

function getNearestAssemblyPoint(fromPoint) {
  const cfg = loadCampusGeoConfig();
  const list = (cfg.points || []).filter((p) => p.type === "assembly");
  if (!list.length) return null;
  let best = list[0];
  let bestDist = haversineMeters(fromPoint, best.point);
  list.slice(1).forEach((item) => {
    const d = haversineMeters(fromPoint, item.point);
    if (d < bestDist) {
      best = item;
      bestDist = d;
    }
  });
  return { point: best, distance: bestDist };
}

function bindSystemModule() {
  async function fetchUsersAndRender() {
    const tbody = $id("userTable");
    const result = await send({
      method: "GET",
      url: $id("baseRole").value + "/api/users?page=1&pageSize=50",
    });
    if (result && result.error) {
      if (tbody) {
        tbody.innerHTML = "<tr><td colspan=\"3\">角色服务不可用：请先启动后端（或检查 baseRole）</td></tr>";
      }
      return;
    }
    const items = result && result.data && result.data.items ? result.data.items : [];
    tbody.innerHTML = "";
    items.forEach((u) => {
      const tr = document.createElement("tr");
      tr.innerHTML = "<td>" + u.username + "</td><td>" + u.email + "</td><td>" + (u.role_id || "-") + "</td>";
      tbody.appendChild(tr);
    });
  }

  async function resolveUserByName(name) {
    const result = await send({
      method: "GET",
      url: $id("baseRole").value + "/api/users?page=1&pageSize=200",
    });
    if (result && result.error) return null;
    const items = result && result.data && result.data.items ? result.data.items : [];
    return items.find((u) => u.username === name) || null;
  }

  fetchUsersAndRender();
  initCampusMaps($id("viewSystem") || document);
  ensureAdminMapEventsBound();
  updateAdminGeoSummary();

  $id("btnUserAdd").addEventListener("click", async () => {
    await send({
      method: "POST",
      url: $id("baseRole").value + "/api/users",
      body: {
        username: $id("uName").value,
        password: $id("uPass").value,
        email: $id("uEmail").value,
        role_id: toUint($id("uRole").value),
      },
    });
    await fetchUsersAndRender();
    updateMetrics();
  });

  $id("btnUserEdit").addEventListener("click", async () => {
    const target = await resolveUserByName($id("uName").value);
    if (!target) {
      logLocal("编辑用户", "未找到对应用户，请先点击角色服务中的用户列表确认");
      return;
    }
    await send({
      method: "PUT",
      url: $id("baseRole").value + "/api/users/" + target.id,
      body: {
        password: $id("uPass").value,
        email: $id("uEmail").value,
        role_id: toUint($id("uRole").value),
        is_active: true,
      },
    });
    await fetchUsersAndRender();
  });

  $id("btnUserDelete").addEventListener("click", async () => {
    const target = await resolveUserByName($id("uName").value);
    if (!target) {
      logLocal("删除用户", "未找到对应用户");
      return;
    }
    await send({
      method: "DELETE",
      url: $id("baseRole").value + "/api/users/" + target.id,
    });
    await fetchUsersAndRender();
    updateMetrics();
  });

  $id("btnUserPerm").addEventListener("click", async () => {
    const target = await resolveUserByName($id("uName").value);
    if (!target) {
      logLocal("用户权限分配", "未找到对应用户");
      return;
    }
    await send({
      method: "POST",
      url: $id("baseRole").value + "/api/users/" + target.id + "/permissions",
      body: { permission: $id("userPerm").value },
    });
  });

  $id("btnRoleCreate").addEventListener("click", async () => {
    await send({
      method: "POST",
      url: $id("baseRole").value + "/api/roles",
      body: parseJSON($id("roleCreatePayload").value, {}),
    });
    updateMetrics();
  });

  $id("btnRoleList").addEventListener("click", async () => {
    await send({
      method: "GET",
      url: $id("baseRole").value + "/api/roles?page=1&pageSize=20",
    });
    updateMetrics();
  });

  $id("btnRoleGet").addEventListener("click", async () => {
    await send({ method: "GET", url: $id("baseRole").value + "/api/roles/" + $id("roleId").value });
  });

  $id("btnRoleUpdate").addEventListener("click", async () => {
    await send({
      method: "PUT",
      url: $id("baseRole").value + "/api/roles/" + $id("roleId").value,
      body: parseJSON($id("roleUpdateBody").value, {}),
    });
  });

  $id("btnRoleDelete").addEventListener("click", async () => {
    await send({ method: "DELETE", url: $id("baseRole").value + "/api/roles/" + $id("roleId").value });
  });

  $id("btnRoleAssignPerm").addEventListener("click", async () => {
    await send({
      method: "POST",
      url: $id("baseRole").value + "/api/roles/" + $id("roleId").value + "/permissions",
      body: { permission_ids: parseJSON($id("rolePermIds").value, []) },
    });
  });

  $id("btnRoleGetPerm").addEventListener("click", async () => {
    await send({ method: "GET", url: $id("baseRole").value + "/api/roles/" + $id("roleId").value + "/permissions" });
  });

  $id("btnRoleExport").addEventListener("click", async () => {
    await send({ method: "GET", url: $id("baseRole").value + "/api/roles/export" });
  });

  $id("btnPermAssign").addEventListener("click", async () => {
    logLocal("权限分配", "当前后端支持角色权限分配与用户权限分配，请使用对应模块按钮。\n权限树会通过 /api/permissions/tree 查看。 ");
    await send({ method: "GET", url: $id("baseRole").value + "/api/permissions/tree" });
  });

  $id("btnPermImport").addEventListener("click", async () => {
    const file = $id("permImport").files[0];
    if (!file) {
      logLocal("导入权限配置", "未选择文件");
      return;
    }
    const text = await file.text();
    const payload = parseJSON(text, { items: [] });
    await send({ method: "POST", url: $id("baseRole").value + "/api/permissions/import", body: payload });
    await send({ method: "GET", url: $id("baseRole").value + "/api/permissions/audit" });
  });

  const btnAdminStartBoundary = $id("btnAdminStartBoundary");
  if (btnAdminStartBoundary) {
    btnAdminStartBoundary.addEventListener("click", () => {
      if (!isAdminUser()) {
        logLocal("校园范围配置", "仅 admin 可编辑校园范围");
        return;
      }
      initCampusMaps($id("sysCampus") || document);
      ensureAdminMapEventsBound();
      mapState.adminBoundaryDraft = [];
      renderAdminBoundaryDraft();
      setAdminEditMode("boundary");
    });
  }

  const btnAdminFinishBoundary = $id("btnAdminFinishBoundary");
  if (btnAdminFinishBoundary) {
    btnAdminFinishBoundary.addEventListener("click", () => {
      if (!isAdminUser()) return;
      if (mapState.adminBoundaryDraft.length < 3) {
        logLocal("校园范围配置", "至少点击 3 个点才能形成校园范围");
        return;
      }
      const cfg = loadCampusGeoConfig();
      cfg.outline = mapState.adminBoundaryDraft.map((p) => clonePoint(p));
      cfg.center = clonePoint(cfg.outline[0]);
      saveCampusGeoConfig(cfg);
      refreshAllCampusOverlays();
      updateAdminGeoSummary();
      setAdminEditMode("none");
      renderAdminBoundaryDraft();
      if (mapState.adminMap) {
        mapState.adminMap.fitBounds(L.latLngBounds(cfg.outline), { padding: [30, 30] });
      }
      logLocal("校园范围配置", "校园范围已更新");
    });
  }

  const btnAdminUndoBoundary = $id("btnAdminUndoBoundary");
  if (btnAdminUndoBoundary) {
    btnAdminUndoBoundary.addEventListener("click", () => {
      if (!isAdminUser()) return;
      mapState.adminBoundaryDraft.pop();
      renderAdminBoundaryDraft();
      logLocal("校园范围配置", "已撤销最后一个边界点");
    });
  }

  const btnAdminClearBoundary = $id("btnAdminClearBoundary");
  if (btnAdminClearBoundary) {
    btnAdminClearBoundary.addEventListener("click", () => {
      if (!isAdminUser()) return;
      const cfg = loadCampusGeoConfig();
      cfg.outline = getDefaultCampusGeoConfig().outline;
      saveCampusGeoConfig(cfg);
      mapState.adminBoundaryDraft = [];
      renderAdminBoundaryDraft();
      refreshAllCampusOverlays();
      updateAdminGeoSummary();
      logLocal("校园范围配置", "已清空并回退校园范围为默认值");
    });
  }

  const btnAdminStartPoint = $id("btnAdminStartPoint");
  if (btnAdminStartPoint) {
    btnAdminStartPoint.addEventListener("click", () => {
      if (!isAdminUser()) {
        logLocal("校园范围配置", "仅 admin 可新增点位");
        return;
      }
      initCampusMaps($id("sysCampus") || document);
      ensureAdminMapEventsBound();
      setAdminEditMode("point");
    });
  }

  const btnAdminStopEdit = $id("btnAdminStopEdit");
  if (btnAdminStopEdit) {
    btnAdminStopEdit.addEventListener("click", () => {
      setAdminEditMode("none");
      mapState.adminBoundaryDraft = [];
      renderAdminBoundaryDraft();
    });
  }

  const btnAdminClearPoints = $id("btnAdminClearPoints");
  if (btnAdminClearPoints) {
    btnAdminClearPoints.addEventListener("click", () => {
      if (!isAdminUser()) return;
      const cfg = loadCampusGeoConfig();
      cfg.points = [];
      saveCampusGeoConfig(cfg);
      refreshAllCampusOverlays();
      updateAdminGeoSummary();
      logLocal("校园范围配置", "已清空全部点位");
    });
  }

  const btnAdminSaveCampusGeo = $id("btnAdminSaveCampusGeo");
  if (btnAdminSaveCampusGeo) {
    btnAdminSaveCampusGeo.addEventListener("click", () => {
      if (!isAdminUser()) return;
      const cfg = saveCampusGeoConfig(loadCampusGeoConfig());
      updateAdminGeoSummary();
      logLocal("校园范围配置保存", cfg);
    });
  }

  const btnAdminApplyCampusGeo = $id("btnAdminApplyCampusGeo");
  if (btnAdminApplyCampusGeo) {
    btnAdminApplyCampusGeo.addEventListener("click", () => {
      refreshAllCampusOverlays();
      updateAdminGeoSummary();
      logLocal("校园范围发布", "已同步到路径/疏散地图");
    });
  }

  const btnAdminExportCampusGeo = $id("btnAdminExportCampusGeo");
  if (btnAdminExportCampusGeo) {
    btnAdminExportCampusGeo.addEventListener("click", () => {
      const cfg = loadCampusGeoConfig();
      const text = JSON.stringify(cfg, null, 2);
      const ta = $id("adminCampusJson");
      if (ta) ta.value = text;
      logLocal("导出校园配置", "已写入文本框，可复制备份");
    });
  }

  const btnAdminImportCampusGeo = $id("btnAdminImportCampusGeo");
  if (btnAdminImportCampusGeo) {
    btnAdminImportCampusGeo.addEventListener("click", () => {
      if (!isAdminUser()) return;
      const ta = $id("adminCampusJson");
      const payload = parseJSON((ta && ta.value) || "", null);
      if (!payload) {
        logLocal("导入校园配置", "JSON 无效，请检查格式");
        return;
      }
      const cfg = saveCampusGeoConfig(payload);
      refreshAllCampusOverlays();
      updateAdminGeoSummary();
      logLocal("导入校园配置", { outline_points: cfg.outline.length, points: cfg.points.length });
    });
  }

  const btnAdminResetCampusGeo = $id("btnAdminResetCampusGeo");
  if (btnAdminResetCampusGeo) {
    btnAdminResetCampusGeo.addEventListener("click", () => {
      if (!isAdminUser()) return;
      const cfg = saveCampusGeoConfig(getDefaultCampusGeoConfig());
      const ta = $id("adminCampusJson");
      if (ta) ta.value = JSON.stringify(cfg, null, 2);
      refreshAllCampusOverlays();
      updateAdminGeoSummary();
      logLocal("校园范围配置", "已恢复默认校园范围与点位");
    });
  }

  const btnAdminFindNearestAssembly = $id("btnAdminFindNearestAssembly");
  if (btnAdminFindNearestAssembly) {
    btnAdminFindNearestAssembly.addEventListener("click", async () => {
      let from = null;
      if (mapState.userLocation) {
        from = mapState.userLocation;
      } else if (mapState.adminMap) {
        const c = mapState.adminMap.getCenter();
        from = [c.lat, c.lng];
      }
      if (!from) {
        logLocal("最近集合点", "没有可用参考点，请先定位或打开地图");
        return;
      }
      const nearest = getNearestAssemblyPoint(from);
      if (!nearest) {
        logLocal("最近集合点", "当前未配置集合点");
        return;
      }
      logLocal("最近集合点", {
        from,
        nearest: nearest.point,
        distance_m: roundMeters(nearest.distance),
      });
    });
  }

  const btnAdminGenerateDraftRoute = $id("btnAdminGenerateDraftRoute");
  if (btnAdminGenerateDraftRoute) {
    btnAdminGenerateDraftRoute.addEventListener("click", () => {
      const cfg = loadCampusGeoConfig();
      const risks = (cfg.points || []).filter((p) => p.type === "risk");
      const assemblies = (cfg.points || []).filter((p) => p.type === "assembly");
      if (!risks.length || !assemblies.length) {
        logLocal("演练疏散草案", "请至少配置 1 个风险点和 1 个集合点");
        return;
      }
      const items = risks.map((r) => {
        const nearest = getNearestAssemblyPoint(r.point);
        return {
          risk: r.name,
          assembly: nearest && nearest.point ? nearest.point.name : "-",
          distance_m: nearest ? roundMeters(nearest.distance) : 0,
        };
      });
      logLocal("演练疏散草案", items);
    });
  }
}

async function updateMetrics() {
  try {
    const navRes = await fetch($id("baseNav").value + "/api/navigation/summary?" + getNavDateQuery(), {
      headers: { "X-User-ID": "1001" },
    });
    const navText = await navRes.text();
    const navData = parseJSON(navText, {});
    const count = navData && navData.data && navData.data.completed_count;
    if (count !== undefined) $id("metricPath").textContent = String(count);
  } catch (_) {}

  try {
    const planRes = await fetch($id("basePlan").value + "/api/plans/search?page=1&page_size=1");
    const planData = parseJSON(await planRes.text(), {});
    const total = planData && planData.data && planData.data.total;
    if (total !== undefined) $id("metricPlan").textContent = String(total);
  } catch (_) {}

  try {
    const roleRes = await fetch($id("baseRole").value + "/api/roles?page=1&pageSize=1");
    const roleData = parseJSON(await roleRes.text(), {});
    const total = roleData && roleData.data && roleData.data.total;
    if (total !== undefined) $id("metricRole").textContent = String(total);
  } catch (_) {}
}

function bindMisc() {
  $id("btnPlanOptimize").addEventListener("click", updateMetrics);
}

wireMenu();
wireSubtabs();
// 登录/注册已停用，不再绑定认证 tab 行为。
initCampusMaps(document.querySelector(".view.active") || document);
// 登录/注册已停用，不再绑定登录事件。
bindPathModule();
bindEmergencyModule();
bindReportModule();
bindMonitorModule();
bindSystemModule();
bindMisc();
initAuth();
updateMetrics();
