import axios from "axios";

const baseURL = (
  import.meta.env.VITE_API_BASE_URL ||
  import.meta.env.VITE_APP_API_BASE_URL ||
  (import.meta.env.DEV ? "http://localhost:8080" : "")
).replace(/\/+$/, "");

const API = axios.create({
  baseURL,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: false,
});

export { API as api };
export default API;
