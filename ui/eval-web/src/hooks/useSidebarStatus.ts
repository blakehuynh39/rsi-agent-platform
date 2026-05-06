import { useEffect, useRef, useState } from "react";
import { api } from "@/lib/api";
import type { StatusResponse } from "@/lib/api";

const POLL_MS = 10_000;

/**
 * Light-weight status poll for the app shell (sidebar). The Status page uses
 * its own faster interval; we keep this slower to avoid duplicate load.
 */
export function useSidebarStatus() {
  const [status, setStatus] = useState<StatusResponse | null>(null);
  const inFlightRef = useRef(false);

  useEffect(() => {
    const load = () => {
      if (inFlightRef.current) return;
      inFlightRef.current = true;
      api
        .getStatus()
        .then(setStatus)
        .catch(() => {})
        .finally(() => {
          inFlightRef.current = false;
        });
    };
    load();
    const id = setInterval(load, POLL_MS);
    return () => clearInterval(id);
  }, []);

  return status;
}
