<template>
  <div
    v-if="hasMusic"
    class="ambient-sound"
    :class="{
      'is-loading': status === 'loading',
      'is-playing': status === 'playing',
      'is-error': status === 'error',
      'has-track': !!displayText,
      'has-lyrics': lyricLines.length > 0 && status === 'playing',
      'is-expanded': expanded,
    }"
    :aria-label="status === 'playing' ? '关闭音乐' : '开启音乐'"
    :aria-pressed="status === 'playing'"
    @click="onToggleExpanded"
  >
    <span class="ambient-bars" aria-hidden="true" @click.stop="togglePlay">
      <span></span>
      <span></span>
      <span></span>
      <span></span>
    </span>
    <span class="ambient-label ellipsis" aria-hidden="true">{{ displayText || '&nbsp;' }}</span>
    <span class="ambient-progress" :class="{ 'is-visible': showProgress }" @click.stop>
      <input
        type="range"
        class="ambient-range"
        min="0"
        :max="duration || 0"
        :value="currentTime"
        step="0.1"
        :disabled="!duration"
        @input="onSeek"
        @mousedown="onSeekStart"
        @mouseup="onSeekEnd"
      />
    </span>

    <audio
      ref="audioRef"
      :src="proxiedUrl"
      preload="none"
      :loop="music.loop"
      crossorigin="anonymous"
      @playing="onPlaying"
      @waiting="onWaiting"
      @pause="onPause"
      @ended="onEnded"
      @error="onError"
      @timeupdate="onTimeUpdate"
      @seeked="onSeeked"
    />
  </div>
</template>

<script setup lang="ts">
import type { MusicItemVO } from "~/types";

interface LyricLine {
  time: number;
  text: string;
}

const props = defineProps<{
  music: MusicItemVO;
  showProgress?: boolean;
}>();

const audioRef = ref<HTMLAudioElement | null>(null);
const status = ref<'idle' | 'loading' | 'playing' | 'error'>('idle');
const audioCtx = ref<AudioContext | null>(null);
const analyser = ref<AnalyserNode | null>(null);
const sourceNode = ref<MediaElementAudioSourceNode | null>(null);
const bars = ref([8, 14, 10, 6]);
const canvasAnimFrame = ref(0);
const lyricFrame = ref(0);
const progressFrame = ref(0);
const lyricLines = ref<LyricLine[]>([]);
const currentLyricIndex = ref(-1);
const loadedLyricUrl = ref('');
const currentTime = ref(0);
const duration = ref(0);
const seeking = ref(false);
const expanded = ref(false);
let expandTimer = 0;

const hasMusic = computed(() => Boolean(props.music?.url));
const proxiedUrl = computed(() => {
  const url = props.music?.url;
  if (!url) return '';
  if (url.startsWith('/api/') || url.startsWith('http://127.0.0.1') || url.startsWith('http://localhost')) {
    return url;
  }
  return `/api/file/proxy?url=${encodeURIComponent(url)}`;
});

const trackLabel = computed(() => {
  const parts: string[] = [];
  if (props.music.title) parts.push(props.music.title);
  if (props.music.artist) parts.push(props.music.artist);
  return parts.join(' · ') || '';
});

const displayText = computed(() => {
  if (lyricLines.value.length > 0 && currentLyricIndex.value >= 0) {
    return lyricLines.value[currentLyricIndex.value].text;
  }
  if (status.value === 'playing' && trackLabel.value) {
    return trackLabel.value;
  }
  return trackLabel.value;
});

// ---- LRC parser (ported from Morpho) ----
const lrcTimeParse = (str: string): number | null => {
  const m = str.match(/(?:(\d+):)?(\d{1,2}):(\d{1,2})(?:[.:](\d{1,3}))?/);
  if (!m) return null;
  const h = Number(m[1] || 0);
  const min = Number(m[2] || 0);
  const s = Number(m[3] || 0);
  const ms = m[4] ? Number('0.' + m[4].padEnd(3, '0').slice(0, 3)) : 0;
  return h * 3600 + min * 60 + s + ms;
};

const parseLrc = (text: string): LyricLine[] => {
  const lines: LyricLine[] = [];
  const noTimeLines: string[] = [];
  text.replace(/\r/g, '').split('\n').map(l => l.trim()).filter(Boolean).forEach(line => {
    const times = Array.from(line.matchAll(/\[([^\]]+)\]/g))
      .map(m => lrcTimeParse(m[1]))
      .filter(t => t !== null);
    const text = line.replace(/\[[^\]]+\]/g, '').trim();
    if (times.length && text) {
      times.forEach(t => lines.push({ time: t, text }));
    } else if (text && !/^(ti|ar|al|by|offset):/i.test(text)) {
      noTimeLines.push(text);
    }
  });
  if (lines.length) {
    return lines.sort((a, b) => a.time - b.time);
  }
  return noTimeLines.map((text, i) => ({ time: i * 5, text }));
};

const loadLyrics = async () => {
  const lyricUrl = props.music.lyricUrl;
  if (!lyricUrl || lyricUrl === loadedLyricUrl.value) return;

  loadedLyricUrl.value = lyricUrl;
  try {
    // /upload/ paths are local static files — don't go through proxy
    const proxyUrl = lyricUrl.startsWith('/upload/') || lyricUrl.startsWith('/api/') || lyricUrl.startsWith('http://127.0.0.1') || lyricUrl.startsWith('http://localhost')
      ? lyricUrl
      : `/api/file/proxy?url=${encodeURIComponent(lyricUrl)}`;
    const resp = await fetch(proxyUrl);
    if (!resp.ok) throw new Error('Failed to load lyrics');
    const text = await resp.text();
    lyricLines.value = parseLrc(text);
    currentLyricIndex.value = -1;
  } catch {
    lyricLines.value = [];
    loadedLyricUrl.value = '';
  }
};

const updateLyric = () => {
  const audio = audioRef.value;
  if (!audio || !lyricLines.value.length) {
    lyricFrame.value = 0;
    return;
  }

  const t = audio.currentTime;
  let idx = currentLyricIndex.value;
  if (idx < 0 || lyricLines.value[idx]?.time > t) {
    idx = 0;
  }
  while (idx < lyricLines.value.length - 1 && lyricLines.value[idx + 1].time <= t + 0.12) {
    idx++;
  }
  if (idx !== currentLyricIndex.value) {
    currentLyricIndex.value = idx;
  }
  lyricFrame.value = requestAnimationFrame(updateLyric);
};

const updateProgress = () => {
  const audio = audioRef.value;
  if (!audio) {
    progressFrame.value = 0;
    return;
  }
  if (!seeking.value) {
    currentTime.value = audio.currentTime;
    duration.value = audio.duration || 0;
  }
  progressFrame.value = requestAnimationFrame(updateProgress);
};

const isBuffered = (audio: HTMLAudioElement, time: number): boolean => {
  for (let i = 0; i < audio.buffered.length; i++) {
    if (time >= audio.buffered.start(i) && time <= audio.buffered.end(i)) {
      return true;
    }
  }
  return false;
};

const onSeek = (e: Event) => {
  const val = parseFloat((e.target as HTMLInputElement).value);
  currentTime.value = val;
};

const onSeekStart = () => {
  seeking.value = true;
};

const onSeekEnd = () => {
  const audio = audioRef.value;
  if (!audio) return;
  seeking.value = false;

  const targetTime = currentTime.value;
  if (!isBuffered(audio, targetTime)) {
    currentTime.value = audio.currentTime;
    return;
  }

  audio.currentTime = targetTime;
};

const onTimeUpdate = () => {
  if (!seeking.value) {
    currentTime.value = audioRef.value?.currentTime || 0;
  }
};

const onSeeked = () => {
  if (lyricFrame.value) {
    cancelAnimationFrame(lyricFrame.value);
  }
  updateLyric();
};

// ---- Mobile expand/collapse ----
const onToggleExpanded = () => {
  if (window.innerWidth >= 640) return;
  expanded.value = !expanded.value;
  if (expanded.value) {
    clearTimeout(expandTimer);
    expandTimer = window.setTimeout(() => {
      expanded.value = false;
    }, 5000);
  } else {
    clearTimeout(expandTimer);
  }
};

// ---- Audio controls ----
const togglePlay = async () => {
  const audio = audioRef.value;
  if (!audio) return;

  if (status.value === 'playing') {
    audio.pause();
    return;
  }

  status.value = 'loading';
  try {
    await loadLyrics();
    await audio.play();
  } catch {
    status.value = 'error';
    setTimeout(() => { status.value = 'idle'; }, 1200);
  }
};

const initAnalyzer = () => {
  const audio = audioRef.value;
  if (!audio || audioCtx.value) return;

  try {
    const AC = (window as any).AudioContext || (window as any).webkitAudioContext;
    if (!AC) return;

    const ctx = new AC() as AudioContext;
    const anl = ctx.createAnalyser();
    anl.fftSize = 128;
    anl.smoothingTimeConstant = 0.74;

    const src = ctx.createMediaElementSource(audio);
    src.connect(anl);
    anl.connect(ctx.destination);

    if (ctx.state === 'suspended') {
      ctx.resume();
    }

    audioCtx.value = ctx;
    analyser.value = anl;
    sourceNode.value = src;
  } catch {
    // AudioContext not supported or already connected
  }
};

const startVisualization = () => {
  if (!analyser.value) return;

  const buffer = new Uint8Array(analyser.value.frequencyBinCount);
  const el = Array.from(document.querySelectorAll('.ambient-sound .ambient-bars span'));

  const tick = () => {
    if (!analyser.value || status.value !== 'playing') {
      canvasAnimFrame.value = 0;
      return;
    }

    analyser.value.getByteFrequencyData(buffer);

    const bands = [
      avg(buffer, 2, 6),
      avg(buffer, 6, 13),
      avg(buffer, 13, 28),
      avg(buffer, 28, 52),
    ];

    bands.forEach((val, i) => {
      const norm = Math.pow(Math.min(val / 255, 1), 0.68);
      const h = 5 + norm * 15;
      bars.value[i] = bars.value[i] * 0.66 + h * 0.34;
      if (el[i]) {
        (el[i] as HTMLElement).style.setProperty('height', `${bars.value[i].toFixed(2)}px`);
      }
    });

    canvasAnimFrame.value = requestAnimationFrame(tick);
  };

  canvasAnimFrame.value = requestAnimationFrame(tick);
};

const avg = (arr: Uint8Array, start: number, end: number) => {
  let sum = 0;
  const count = Math.min(end, arr.length) - start;
  for (let i = start; i < Math.min(end, arr.length); i++) sum += arr[i];
  return sum / Math.max(1, count);
};

const onPlaying = () => {
  status.value = 'playing';
  duration.value = audioRef.value?.duration || 0;
  initAnalyzer();
  startVisualization();
  updateLyric();
  updateProgress();
};

const onWaiting = () => {
  status.value = 'loading';
};

const stopAll = () => {
  if (canvasAnimFrame.value) {
    cancelAnimationFrame(canvasAnimFrame.value);
    canvasAnimFrame.value = 0;
  }
  if (lyricFrame.value) {
    cancelAnimationFrame(lyricFrame.value);
    lyricFrame.value = 0;
  }
  if (progressFrame.value) {
    cancelAnimationFrame(progressFrame.value);
    progressFrame.value = 0;
  }
};

const onPause = () => {
  status.value = 'idle';
  stopAll();
};

const onEnded = () => {
  status.value = 'idle';
  stopAll();
};

const onError = () => {
  status.value = 'error';
  stopAll();
};

watch(
  () => props.music.url,
  () => {
    status.value = 'idle';
    stopAll();
    audioCtx.value = null;
    analyser.value = null;
    sourceNode.value = null;
    loadedLyricUrl.value = '';
    lyricLines.value = [];
    currentLyricIndex.value = -1;
    currentTime.value = 0;
    duration.value = 0;
  }
);
</script>

<style scoped>
.ambient-sound {
  --ambient-bar-1: 9px;
  --ambient-bar-2: 15px;
  --ambient-bar-3: 11px;
  --ambient-bar-4: 7px;
  --ambient-opacity-1: 0.78;
  --ambient-opacity-2: 1;
  --ambient-opacity-3: 0.86;
  --ambient-opacity-4: 0.68;
  width: max-content;
  max-width: min(260px, 100vw - 150px);
  height: var(--bar-item-size, 32px);
  color: inherit;
  opacity: 0.88;
  transition: color 0.22s, opacity 0.22s, transform 0.18s;
  border-radius: 4px;
  display: flex;
  justify-content: flex-start;
  align-items: center;
  gap: 6px;
  padding: 0 8px 0 0;
  cursor: pointer;
  user-select: none;
  -webkit-user-select: none;
}

.ambient-sound:focus-visible {
  outline: 2px solid rgba(255, 255, 255, 0.5);
  outline-offset: 2px;
}

.ambient-sound.is-playing {
  opacity: 1;
}

.ambient-sound.is-loading {
  opacity: 0.88;
}

.ambient-sound.is-error {
  opacity: 0.45;
}

.ambient-bars {
  width: var(--bar-item-size, 32px);
  height: var(--bar-item-size, 32px);
  flex: none;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 2px;
}

.ambient-bars span {
  transform-origin: 50%;
  background-color: currentColor;
  border-radius: 999px;
  width: 2px;
  height: 12px;
  display: block;
  transition: height 0.22s, opacity 0.22s, transform 0.22s;
}

.ambient-bars span:first-child {
  height: var(--ambient-bar-1);
  opacity: var(--ambient-opacity-1);
}

.ambient-bars span:nth-child(2) {
  height: var(--ambient-bar-2);
  opacity: var(--ambient-opacity-2);
}

.ambient-bars span:nth-child(3) {
  height: var(--ambient-bar-3);
  opacity: var(--ambient-opacity-3);
}

.ambient-bars span:nth-child(4) {
  height: var(--ambient-bar-4);
  opacity: var(--ambient-opacity-4);
}

/* Not playing: show static bars with focus */
.ambient-bars span:first-child,
.ambient-sound:not(.is-playing):focus-visible .ambient-bars span:first-child {
  height: 12px;
}

.ambient-bars span:nth-child(2),
.ambient-sound:not(.is-playing):focus-visible .ambient-bars span:nth-child(2) {
  height: 17px;
}

.ambient-bars span:nth-child(3),
.ambient-sound:not(.is-playing):focus-visible .ambient-bars span:nth-child(3) {
  height: 14px;
}

.ambient-bars span:nth-child(4),
.ambient-sound:not(.is-playing):focus-visible .ambient-bars span:nth-child(4) {
  height: 10px;
}

/* Loading animation */
@keyframes ambient-load {
  0%, to {
    opacity: 0.45;
    transform: translateY(0);
  }
  50% {
    opacity: 1;
    transform: translateY(-3px);
  }
}

.ambient-sound.is-loading .ambient-bars span {
  animation: 0.72s ease-in-out infinite ambient-load;
}

.ambient-sound.is-loading .ambient-bars span:nth-child(2) {
  animation-delay: 80ms;
}

.ambient-sound.is-loading .ambient-bars span:nth-child(3) {
  animation-delay: 0.16s;
}

.ambient-sound.is-loading .ambient-bars span:nth-child(4) {
  animation-delay: 0.24s;
}

/* Playing but not visualized (fallback) */
@keyframes ambient-breathe {
  0%, to {
    transform: scaleY(0.72);
  }
  50% {
    transform: scaleY(1.1);
  }
}

.ambient-sound.is-playing:not(.is-visualized) .ambient-bars span {
  animation: 3.8s ease-in-out infinite ambient-breathe;
}

.ambient-sound.is-playing:not(.is-visualized) .ambient-bars span:nth-child(2) {
  animation-delay: -1.1s;
}

.ambient-sound.is-playing:not(.is-visualized) .ambient-bars span:nth-child(3) {
  animation-delay: -2.2s;
}

.ambient-sound.is-playing:not(.is-visualized) .ambient-bars span:nth-child(4) {
  animation-delay: -3s;
}

/* Label */
.ambient-label {
  color: currentColor;
  max-width: 0;
  font-size: 11px;
  line-height: 18px;
  opacity: 0;
  pointer-events: none;
  white-space: nowrap;
  overflow: hidden;
  transition: max-width 0.28s cubic-bezier(0.22, 1, 0.36, 1), opacity 0.22s, transform 0.22s;
  transform: translateX(-4px);
}

.ambient-sound.has-track .ambient-label:not(:empty),
.ambient-sound:hover .ambient-label:not(:empty),
.ambient-sound:focus-visible .ambient-label:not(:empty),
.ambient-sound.is-expanded .ambient-label:not(:empty) {
  opacity: 0.8;
  max-width: 150px;
  transform: translateX(0);
}

.ambient-sound.has-lyrics .ambient-label:not(:empty) {
  opacity: 0.88;
}

/* Progress bar */
.ambient-progress {
  width: 0;
  opacity: 0;
  overflow: hidden;
  transition: width 0.28s cubic-bezier(0.22, 1, 0.36, 1), opacity 0.22s, margin 0.22s;
  display: flex;
  align-items: center;
}

.ambient-sound:hover .ambient-progress,
.ambient-sound:focus-visible .ambient-progress,
.ambient-progress.is-visible,
.ambient-sound.is-expanded .ambient-progress {
  width: 72px;
  opacity: 1;
}

.ambient-range {
  width: 100%;
  margin: 0;
  appearance: none;
  -webkit-appearance: none;
  background: transparent;
  cursor: pointer;
}

.ambient-range::-webkit-slider-runnable-track {
  height: 3px;
  background: rgb(255 255 255 / 0.35);
  border-radius: 999px;
}

.ambient-range::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 10px;
  height: 10px;
  margin-top: -3.5px;
  border-radius: 999px;
  border: none;
  background: rgb(255 255 255 / 0.92);
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.3);
}

.ambient-range::-moz-range-track {
  height: 3px;
  background: rgb(255 255 255 / 0.35);
  border: none;
  border-radius: 999px;
}

.ambient-range::-moz-range-thumb {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  border: none;
  background: rgb(255 255 255 / 0.92);
  box-shadow: 0 1px 3px rgb(0 0 0 / 0.3);
}

.ambient-range:disabled {
  opacity: 0.35;
  cursor: default;
}
</style>
