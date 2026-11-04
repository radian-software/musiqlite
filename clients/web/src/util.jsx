import { use } from "react";

export const fetchJSON = async (urlpath) => {
  if (!urlpath.startsWith("/"))
    throw new Error(`fetchJSON urlpath does not start with slash`);
  const resp = await fetch("http://localhost:8080" + urlpath);
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
export const useFetchJSON = (urlpath) => {
  if (!useFetchCache[urlpath]) {
    useFetchCache[urlpath] = fetchJSON(urlpath);
  }
  return use(useFetchCache[urlpath]);
};
