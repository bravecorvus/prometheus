import { useRef, useState } from "react";
import { type State, patchSettings, uploadAudio } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Upload, Mail, Lightbulb, Speaker } from "lucide-react";
import { toast } from "sonner";

interface Props {
  state: State;
  onChange: (next: State) => void;
}

export function SettingsTab({ state, onChange }: Props) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [emailDraft, setEmailDraft] = useState(state.email);
  const emailDirty = emailDraft !== state.email;

  async function applySettings(patch: Partial<State>) {
    const previous = state;
    onChange({ ...state, ...patch });
    try {
      await patchSettings(patch);
    } catch (err) {
      onChange(previous);
      toast.error("Failed to save", { description: String(err) });
    }
  }

  async function onFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    if (!file.type.startsWith("audio/")) {
      toast.error("Audio files only");
      return;
    }
    try {
      const { trackInfo } = await uploadAudio(file);
      onChange({ ...state, trackInfo });
      toast.success("Track uploaded", { description: trackInfo });
    } catch (err) {
      toast.error("Upload failed", { description: String(err) });
    } finally {
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <Upload className="size-4" /> Alarm Audio
          </CardTitle>
          <CardDescription className="truncate">
            {state.trackInfo || "No track uploaded"}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <input
            ref={fileInputRef}
            type="file"
            accept="audio/*"
            className="hidden"
            onChange={onFile}
          />
          <Button
            variant="secondary"
            onClick={() => fileInputRef.current?.click()}
            className="w-full"
          >
            Replace track
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <Lightbulb className="size-4" /> LED
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <Label htmlFor="enable-led">Enable LED</Label>
            <Switch
              id="enable-led"
              checked={state.enableLed}
              onCheckedChange={(v) => applySettings({ enableLed: v })}
            />
          </div>
          {state.enableLed && (
            <div className="flex items-center gap-2">
              <Input
                type="color"
                value={state.colors || "#ffffff"}
                onChange={(e) => applySettings({ colors: e.target.value })}
                className="h-10 w-16 cursor-pointer p-1"
              />
              <span className="font-mono text-sm text-muted-foreground">
                {state.colors || "#ffffff"}
              </span>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <Mail className="size-4" /> IP-Change Email
          </CardTitle>
          <CardDescription>
            Get notified when this Pi's LAN address changes.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <Label htmlFor="enable-email">Enable</Label>
            <Switch
              id="enable-email"
              checked={state.enableEmail}
              onCheckedChange={(v) => applySettings({ enableEmail: v })}
            />
          </div>
          {state.enableEmail && (
            <div className="flex items-center gap-2">
              <Input
                type="email"
                placeholder="you@example.com"
                value={emailDraft}
                onChange={(e) => setEmailDraft(e.target.value)}
              />
              <Button
                size="sm"
                disabled={!emailDirty}
                onClick={() => applySettings({ email: emailDraft })}
              >
                Save
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <Speaker className="size-4" /> Audio Output
          </CardTitle>
          <CardDescription>
            Route playback through an external USB sound card.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <Label htmlFor="custom-soundcard">Custom Sound Card</Label>
            <Switch
              id="custom-soundcard"
              checked={state.customSoundCard}
              onCheckedChange={(v) => applySettings({ customSoundCard: v })}
            />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
