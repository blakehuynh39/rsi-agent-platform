import { useCallback, useEffect, useState } from "react";
import { Clock } from "lucide-react";
import { Badge } from "@nous-research/ui/ui/components/badge";
import { Spinner } from "@nous-research/ui/ui/components/spinner";
import { H2 } from "@/components/NouiTypography";
import { api } from "@/lib/api";
import type { CronJob } from "@/lib/api";
import { useToast } from "@/hooks/useToast";
import { Toast } from "@/components/Toast";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useI18n } from "@/i18n";
import { PluginSlot } from "@/plugins";

function formatTime(iso?: string | null): string {
  if (!iso) return "—";
  const d = new Date(iso);
  return d.toLocaleString();
}

const STATUS_TONE: Record<string, "success" | "warning" | "destructive"> = {
  enabled: "success",
  scheduled: "success",
  paused: "warning",
  error: "destructive",
  completed: "destructive",
};

export default function CronPage() {
  const [jobs, setJobs] = useState<CronJob[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast, showToast } = useToast();
  const { t } = useI18n();

  const loadJobs = useCallback(() => {
    api
      .getCronJobs()
      .then(setJobs)
      .catch(() => showToast(t.common.loading, "error"))
      .finally(() => setLoading(false));
  }, [showToast, t.common.loading]);

  useEffect(() => {
    loadJobs();
  }, [loadJobs]);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24">
        <Spinner className="text-2xl text-primary" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <PluginSlot name="cron:top" />
      <Toast toast={toast} />

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Clock className="h-4 w-4" />
            RSI cron state
          </CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          Hermes local cron mutations are disabled here. This page reflects the
          RSI platform scheduler and proposal-slot state exposed by the Go API.
        </CardContent>
      </Card>

      <div className="flex flex-col gap-3">
        <H2
          variant="sm"
          className="flex items-center gap-2 text-muted-foreground"
        >
          <Clock className="h-4 w-4" />
          {t.cron.scheduledJobs} ({jobs.length})
        </H2>

        {jobs.length === 0 && (
          <Card>
            <CardContent className="py-8 text-center text-sm text-muted-foreground">
              {t.cron.noJobs}
            </CardContent>
          </Card>
        )}

        {jobs.map((job) => (
          <Card key={job.id}>
            <CardContent className="flex items-center gap-4 py-4">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className="font-medium text-sm truncate">
                    {job.name ||
                      job.prompt.slice(0, 60) +
                        (job.prompt.length > 60 ? "..." : "")}
                  </span>
                  <Badge tone={STATUS_TONE[job.state] ?? "secondary"}>
                    {job.state}
                  </Badge>
                  {job.deliver && job.deliver !== "local" && (
                    <Badge tone="outline">{job.deliver}</Badge>
                  )}
                </div>
                {job.name && (
                  <p className="text-xs text-muted-foreground truncate mb-1">
                    {job.prompt.slice(0, 100)}
                    {job.prompt.length > 100 ? "..." : ""}
                  </p>
                )}
                <div className="flex items-center gap-4 text-xs text-muted-foreground">
                  <span className="font-mono">{job.schedule_display}</span>
                  <span>
                    {t.cron.last}: {formatTime(job.last_run_at)}
                  </span>
                  <span>
                    {t.cron.next}: {formatTime(job.next_run_at)}
                  </span>
                </div>
                {job.last_error && (
                  <p className="text-xs text-destructive mt-1">
                    {job.last_error}
                  </p>
                )}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <PluginSlot name="cron:bottom" />
    </div>
  );
}
