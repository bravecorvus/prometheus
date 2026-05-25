// Thin fetch wrapper around the backend's /api/* endpoints. Callers do
// optimistic UI updates and use these for the network half — non-2xx
// responses throw so the caller can roll back local state.

export interface Alarm {
  name: string;
  time: string;
  sound: boolean;
  vibration: boolean;
}

export interface State {
  demo: boolean;
  alarms: Alarm[];
  email: string;
  enableEmail: boolean;
  enableLed: boolean;
  customSoundCard: boolean;
  colors: string;
  trackInfo: string;
}

export type AlarmPatch = Partial<Pick<Alarm, "time" | "sound" | "vibration">>;

export type SettingsPatch = Partial<{
  email: string;
  enableEmail: boolean;
  enableLed: boolean;
  customSoundCard: boolean;
  colors: string;
}>;

async function req(input: RequestInfo, init?: RequestInit) {
  const r = await fetch(input, init);
  if (!r.ok) throw new Error(`${r.status} ${r.statusText}`);
  return r;
}

export async function getState(): Promise<State> {
  const r = await req("/api/state");
  return r.json();
}

export async function patchAlarm(name: string, patch: AlarmPatch): Promise<Alarm> {
  const r = await req(`/api/alarms/${encodeURIComponent(name)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(patch),
  });
  return r.json();
}

export async function snooze(): Promise<void> {
  await req("/api/snooze", { method: "POST" });
}

export async function patchSettings(patch: SettingsPatch): Promise<void> {
  await req("/api/settings", {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(patch),
  });
}

export async function uploadAudio(file: File): Promise<{ trackInfo: string }> {
  const fd = new FormData();
  fd.append("audio", file, file.name);
  const r = await req("/api/upload", { method: "POST", body: fd });
  return r.json();
}
