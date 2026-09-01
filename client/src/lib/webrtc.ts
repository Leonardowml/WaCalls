import { float32ToInt16LE, int16LEToFloat32 } from "./pcm";
import {
  CAPTURE_PROCESSOR_NAME,
  CAPTURE_WORKLET_URL,
  PLAYBACK_PROCESSOR_NAME,
  PLAYBACK_WORKLET_URL,
  SAMPLE_RATE,
} from "../constants/audio";

// O audio vai e volta pela mesma conexao HTTPS que serve esta pagina.
// A versao original usava WebRTC, que exige portas UDP proprias e nao
// atravessa proxy — ver cmd/server/wsbridge.go.

export type OpenCall = {
  micStream: MediaStream;
  remoteStream: MediaStream | null;
  close: () => void;
};

const audioSocketURL = (sid: string, callId: string): string => {
  const scheme = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${scheme}//${window.location.host}/api/sessions/${sid}/calls/${callId}/audio`;
};

const waitOpen = (ws: WebSocket): Promise<void> =>
  new Promise<void>((resolve, reject) => {
    const done = () => {
      ws.removeEventListener("open", onOpen);
      ws.removeEventListener("error", onFail);
      ws.removeEventListener("close", onFail);
    };
    const onOpen = () => {
      done();
      resolve();
    };
    const onFail = () => {
      done();
      reject(new Error("nao foi possivel abrir o canal de audio"));
    };
    ws.addEventListener("open", onOpen);
    ws.addEventListener("error", onFail);
    ws.addEventListener("close", onFail);
  });

export const openCall = async (
  sid: string,
  callId: string,
  micDeviceId: string | null,
): Promise<OpenCall> => {
  const micStream = await navigator.mediaDevices.getUserMedia({
    audio: micDeviceId ? { deviceId: { exact: micDeviceId } } : true,
  });

  const ws = new WebSocket(audioSocketURL(sid, callId));
  ws.binaryType = "arraybuffer";
  try {
    await waitOpen(ws);
  } catch (err) {
    micStream.getTracks().forEach((t) => t.stop());
    throw err;
  }

  const ctx = new AudioContext({ sampleRate: SAMPLE_RATE });
  await ctx.audioWorklet.addModule(CAPTURE_WORKLET_URL);
  await ctx.audioWorklet.addModule(PLAYBACK_WORKLET_URL);
  await ctx.resume();

  const micSource = ctx.createMediaStreamSource(micStream);
  const captureNode = new AudioWorkletNode(ctx, CAPTURE_PROCESSOR_NAME);
  captureNode.port.onmessage = (e: MessageEvent<Float32Array>) => {
    if (ws.readyState === WebSocket.OPEN) ws.send(float32ToInt16LE(e.data));
  };
  micSource.connect(captureNode);
  captureNode.connect(ctx.destination);

  const playbackNode = new AudioWorkletNode(ctx, PLAYBACK_PROCESSOR_NAME);
  const streamDest = ctx.createMediaStreamDestination();
  playbackNode.connect(streamDest);
  ws.onmessage = (e: MessageEvent<ArrayBuffer>) => {
    if (e.data instanceof ArrayBuffer) {
      playbackNode.port.postMessage(int16LEToFloat32(e.data));
    }
  };

  return {
    micStream,
    remoteStream: streamDest.stream,
    close: () => {
      try {
        micStream.getTracks().forEach((t) => t.stop());
      } catch {}
      try {
        ctx.close();
      } catch {}
      try {
        ws.close();
      } catch {}
    },
  };
};
