import { useState } from "react";
import { type Alarm, patchAlarm } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Volume2, Vibrate } from "lucide-react";
import { toast } from "sonner";

interface Props {
  index: number;
  alarm: Alarm;
  onChange: (next: Alarm) => void;
}

export function AlarmCard({ index, alarm, onChange }: Props) {
  const [draftTime, setDraftTime] = useState(alarm.time);
  const dirty = draftTime !== alarm.time;

  async function commit<K extends keyof Alarm>(key: K, value: Alarm[K]) {
    const previous = alarm[key];
    onChange({ ...alarm, [key]: value });
    try {
      await patchAlarm(alarm.name, { [key]: value } as never);
    } catch (err) {
      onChange({ ...alarm, [key]: previous });
      toast.error("Failed to save", { description: String(err) });
    }
  }

  async function saveTime() {
    if (!dirty) return;
    await commit("time", draftTime);
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">Alarm {index + 1}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-2">
          <Input
            type="time"
            value={draftTime}
            onChange={(e) => setDraftTime(e.target.value)}
            className="text-xl tabular-nums"
          />
          <Button onClick={saveTime} disabled={!dirty} size="sm">
            Set
          </Button>
        </div>

        <div className="flex items-center justify-between">
          <Label
            htmlFor={`${alarm.name}-sound`}
            className="flex items-center gap-2"
          >
            <Volume2 className="size-4 text-muted-foreground" /> Sound
          </Label>
          <Switch
            id={`${alarm.name}-sound`}
            checked={alarm.sound}
            onCheckedChange={(v) => commit("sound", v)}
          />
        </div>

        <div className="flex items-center justify-between">
          <Label
            htmlFor={`${alarm.name}-vib`}
            className="flex items-center gap-2"
          >
            <Vibrate className="size-4 text-muted-foreground" /> Vibration
          </Label>
          <Switch
            id={`${alarm.name}-vib`}
            checked={alarm.vibration}
            onCheckedChange={(v) => commit("vibration", v)}
          />
        </div>
      </CardContent>
    </Card>
  );
}
