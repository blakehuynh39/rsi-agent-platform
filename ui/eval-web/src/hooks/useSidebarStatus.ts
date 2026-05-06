import { useEffect, useRef, useState } from "react";
import { api } from "@/lib/api";
import type { StatusResponse } from "@/lib/api";

const POLL_MS = 10_000;
let cachedStatus: StatusResponse | null = null;
let inFlight = false;
let pollID: number | null = null;
const listeners = new Set<(status: StatusResponse | null) => void>();

function notify(status: StatusResponse | null) {
  for (const listener of listeners) {
    listener(status);
  }
}

function loadSidebarStatus() {
  if (inFlight) return;
  inFlight = true;
  api
    .getStatus()
    .then((status) => {
      cachedStatus = status;
      notify(status);
    })
    .catch(() => {})
    .finally(() => {
      inFlight = false;
    });
}

function startPolling() {
  loadSidebarStatus();
  if (pollID === null) {
    pollID = window.setInterval(loadSidebarStatus, POLL_MS);
  }
}

function stopPollingIfUnused() {
  if (listeners.size > 0 || pollID === null) return;
  window.clearInterval(pollID);
  pollID = null;
}

/**
 * Light-weight status poll for the app shell (sidebar). The Status page uses
 * its own faster interval; we keep this slower to avoid duplicate load.
 */
export function useSidebarStatus() {
  const [status, setStatus] = useState<StatusResponse | null>(cachedStatus);
  const listenerRef = useRef(setStatus);
  listenerRef.current = setStatus;

  useEffect(() => {
    const listener = (next: StatusResponse | null) => {
      listenerRef.current(next);
    };
    listeners.add(listener);
    listener(cachedStatus);
    startPolling();
    return () => {
      listeners.delete(listener);
      stopPollingIfUnused();
    };
  }, []);

  return status;
}
