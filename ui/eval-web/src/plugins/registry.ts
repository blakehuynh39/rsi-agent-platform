/**
 * Dashboard Plugin SDK + Registry
 *
 * Plugins call window.__HERMES_PLUGINS__.register(name, Component)
 * to register their tab component.
 */

import type React from "react";
import { registerSlot } from "./slots";

// ---------------------------------------------------------------------------
// Plugin registry — plugins call register() to add their component.
// ---------------------------------------------------------------------------

type RegistryListener = () => void;

const _registered: Map<string, React.ComponentType> = new Map();
const _loadErrors: Map<string, string> = new Map();
const _listeners: Set<RegistryListener> = new Set();

function _notify() {
  for (const fn of _listeners) {
    try { fn(); } catch { /* ignore */ }
  }
}

/** Re-run registry subscribers (e.g. after a plugin script onload, or dev HMR re-inject). */
export function notifyPluginRegistry() {
  _notify();
}

/** Register a plugin component. Called by plugin JS bundles. */
function registerPlugin(name: string, component: React.ComponentType) {
  _loadErrors.delete(name);
  _registered.set(name, component);
  _notify();
}

/** Get a registered component by plugin name. */
export function getPluginComponent(name: string): React.ComponentType | undefined {
  return _registered.get(name);
}

export function getPluginLoadError(name: string): string | undefined {
  return _loadErrors.get(name);
}

export function setPluginLoadError(name: string, message: string) {
  _loadErrors.set(name, message);
  _notify();
}

/** Subscribe to registry changes (returns unsubscribe fn). */
export function onPluginRegistered(fn: RegistryListener): () => void {
  _listeners.add(fn);
  return () => _listeners.delete(fn);
}

/** Get current count of registered plugins. */
export function getRegisteredCount(): number {
  return _registered.size;
}

// ---------------------------------------------------------------------------
// Expose plugin registry and defer the heavier SDK helpers.
// ---------------------------------------------------------------------------

declare global {
  interface Window {
    __HERMES_PLUGIN_SDK__: unknown;
    __HERMES_PLUGINS__: {
      register: typeof registerPlugin;
      registerSlot: typeof registerSlot;
    };
  }
}

let pluginSDKReady: Promise<void> | null = null;

export function exposePluginSDK(): Promise<void> {
  window.__HERMES_PLUGINS__ = {
    register: registerPlugin,
    registerSlot,
  };

  pluginSDKReady ??= import("./sdk").then(({ hydratePluginSDK }) => {
    hydratePluginSDK();
  });
  return pluginSDKReady;
}

export function ensurePluginSDK(): Promise<void> {
  return exposePluginSDK();
}
