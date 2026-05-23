import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { Badge } from "@nous-research/ui/ui/components/badge";
import { Button } from "@nous-research/ui/ui/components/button";
import { Select, SelectOption } from "@nous-research/ui/ui/components/select";
import { Tabs, TabsList, TabsTrigger } from "@nous-research/ui/ui/components/tabs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { api, fetchJSON } from "@/lib/api";
import { cn, isoTimeAgo, timeAgo } from "@/lib/utils";
import { useI18n } from "@/i18n";
import { PluginSlot } from "./slots";

export function hydratePluginSDK() {
  window.__HERMES_PLUGIN_SDK__ = {
    React,
    hooks: {
      useState,
      useEffect,
      useCallback,
      useMemo,
      useRef,
      useContext,
      createContext,
    },
    api,
    fetchJSON,
    components: {
      Card,
      CardHeader,
      CardTitle,
      CardContent,
      Badge,
      Button,
      Input,
      Label,
      Select,
      SelectOption,
      Separator,
      Tabs,
      TabsList,
      TabsTrigger,
      PluginSlot,
    },
    utils: { cn, timeAgo, isoTimeAgo },
    useI18n,
  };
}
