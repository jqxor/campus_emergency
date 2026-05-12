const CACHE_VERSION = "campus-map-tile-v2";
const TILE_HOSTS = [
  "webst01.is.autonavi.com",
  "webst02.is.autonavi.com",
  "webst03.is.autonavi.com",
  "webst04.is.autonavi.com",
];

self.addEventListener("install", (event) => {
  event.waitUntil(self.skipWaiting());
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(
        keys
          .filter((key) => key.startsWith("campus-map-tile-") && key !== CACHE_VERSION)
          .map((key) => caches.delete(key))
      )
    ).then(() => self.clients.claim())
  );
});

function isTileRequest(requestUrl) {
  try {
    const url = new URL(requestUrl);
    return TILE_HOSTS.includes(url.hostname) && url.pathname.includes("/appmaptile");
  } catch (_) {
    return false;
  }
}

self.addEventListener("fetch", (event) => {
  const req = event.request;

  if (req.method !== "GET") {
    return;
  }

  if (!isTileRequest(req.url)) {
    return;
  }

  event.respondWith(
    caches.open(CACHE_VERSION).then(async (cache) => {
      const cached = await cache.match(req);
      if (cached) {
        return cached;
      }

      try {
        const networkResp = await fetch(req);
        if (networkResp && (networkResp.ok || networkResp.type === "opaque")) {
          cache.put(req, networkResp.clone()).catch(() => {});
        }
        return networkResp;
      } catch (_) {
        return new Response("", { status: 504, statusText: "Tile cache miss" });
      }
    })
  );
});
