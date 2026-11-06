import { use, useState } from "react";

export const fetchJSON = async (urlpath, params) => {
  let method = "GET";
  if (/^[A-Z]+ /.test(urlpath)) {
    const idx = urlpath.indexOf(" ");
    [method, urlpath] = [
      urlpath.slice(0, idx),
      urlpath.slice(idx + 1, urlpath.length),
    ];
  }
  if (!urlpath.startsWith("/"))
    throw new Error(`fetchJSON urlpath does not start with slash`);
  let url = "http://localhost:8080" + urlpath;
  if (params) {
    const query = new URLSearchParams();
    for (let [key, vals] of Object.entries(params)) {
      if (!Array.isArray(vals)) {
        vals = [vals];
      }
      for (const val of vals) {
        query.append(key, val);
      }
    }
    url += "?" + query.toString();
  }
  const resp = await fetch(url, { method });
  if (resp.status === 204) {
    return null;
  }
  let data;
  try {
    data = await resp.json();
  } catch (err) {
    throw new Error(`HTTP ${resp.status} ${resp.statusText}`);
  }
  if (data.internal_error) {
    throw new Error(data.internal_error);
  }
  if (!resp.ok) {
    throw new Error(`HTTP ${resp.status} ${resp.statusText}`);
  }
  return data;
};

const useFetchCache = {};
export const useFetchJSON = (urlpath, params) => {
  const cacheKey = urlpath + "?" + JSON.stringify(params);
  if (!useFetchCache[cacheKey]) {
    useFetchCache[cacheKey] = fetchJSON(urlpath, params);
  }
  return use(useFetchCache[cacheKey]);
};

export const parseBlob = (blob) => {
  const [_, size, digest, propsList] = blob.split(":");
  const props = {};
  for (const propEntry of propsList.split(",")) {
    const [key, val] = propEntry.split("=");
    props[key] = val;
  }
  return { size, digest, props };
};

export const formatDuration = (totalSeconds) => {
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds / 60) % 60);
  const seconds = Math.floor(totalSeconds % 60);
  const parts = [];
  if (hours > 0) {
    parts.push(hours);
    parts.push(("00" + minutes).slice(-2));
  } else {
    parts.push(minutes);
  }
  parts.push(("00" + seconds).slice(-2));
  return parts.join(":");
};

// https://github.com/react/react/issues/14981#issuecomment-468460187
export const useErrorHandling = () => {
  const [_state, setState] = useState();
  return (handler) =>
    async (...args) => {
      try {
        return await handler(...args);
      } catch (err) {
        setState(() => {
          throw err;
        });
      }
    };
};
