<template>
  <div
    v-if="hasMusic"
    class="top-music"
    :class="{ 'is-playing': isPlaying, 'is-expanded': isPlaying || isSeeking }"
  >
    <button
      type="button"
      class="top-music__entry"
      :aria-label="isPlaying ? '暂停音乐' : '播放音乐'"
      title="音乐播放器"
      @click="togglePlay"
    >
      <UIcon
        :name="isPlaying ? 'i-carbon-pause-filled' : 'i-carbon-play-filled'"
        class="top-music__icon"
      />
    </button>

    <Transition name="top-music-track">
      <div v-if="isPlaying || isSeeking" class="top-music__track-wrap">
        <input
          v-model="progress"
          class="top-music__range"
          type="range"
          min="0"
          :max="duration || 0"
          step="0.1"
          @input="seek"
          @mousedown="isSeeking = true"
          @mouseup="isSeeking = false"
          @touchstart="isSeeking = true"
          @touchend="isSeeking = false"
        />
      </div>
    </Transition>

    <audio
      ref="audioRef"
      :src="music.url"
      preload="metadata"
      @play="isPlaying = true"
      @pause="onPause"
      @loadedmetadata="onLoadedMetadata"
      @timeupdate="onTimeUpdate"
      @ended="onPause"
    />
  </div>
</template>

<script setup lang="ts">
import type { MusicItemVO } from "~/types";

const props = defineProps<{
  music: MusicItemVO;
}>();

const audioRef = ref<HTMLAudioElement | null>(null);
const isPlaying = ref(false);
const isSeeking = ref(false);
const duration = ref(0);
const progress = ref(0);

const hasMusic = computed(() => Boolean(props.music?.url));

const togglePlay = async () => {
  if (!audioRef.value || !hasMusic.value) {
    return;
  }

  if (audioRef.value.paused) {
    try {
      await audioRef.value.play();
    } catch {
      isPlaying.value = false;
    }
  } else {
    audioRef.value.pause();
  }
};

const onLoadedMetadata = () => {
  duration.value = audioRef.value?.duration || 0;
};

const onTimeUpdate = () => {
  if (!isSeeking.value) {
    progress.value = audioRef.value?.currentTime || 0;
  }
};

const seek = () => {
  if (!audioRef.value) {
    return;
  }
  audioRef.value.currentTime = Number(progress.value);
};

const onPause = () => {
  isPlaying.value = false;
  isSeeking.value = false;
};

watch(
  () => props.music.url,
  () => {
    isPlaying.value = false;
    isSeeking.value = false;
    progress.value = 0;
    duration.value = 0;
    if (audioRef.value) {
      audioRef.value.pause();
      audioRef.value.load();
    }
  },
);
</script>

<style scoped>
.top-music {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  color: rgb(255 255 255 / 0.9);
}

.top-music__entry {
  width: 1.2rem;
  height: 1.2rem;
  border: none;
  background: transparent;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: inherit;
  opacity: 0.92;
  transition: opacity 0.18s ease;
}

.top-music__entry:hover {
  opacity: 1;
}

.top-music__icon {
  width: 0.95rem;
  height: 0.95rem;
}

.top-music__track-wrap {
  width: 5.5rem;
  display: inline-flex;
  align-items: center;
}

.top-music__range {
  width: 100%;
  margin: 0;
  appearance: none;
  -webkit-appearance: none;
  background: transparent;
}

.top-music__range::-webkit-slider-runnable-track {
  height: 2px;
  background: rgb(255 255 255 / 0.42);
  border-radius: 999px;
}

.top-music__range::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 7px;
  height: 7px;
  margin-top: -2.5px;
  border-radius: 999px;
  border: none;
  background: rgb(255 255 255 / 0.92);
}

.top-music__range::-moz-range-track {
  height: 2px;
  background: rgb(255 255 255 / 0.42);
  border: none;
  border-radius: 999px;
}

.top-music__range::-moz-range-thumb {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  border: none;
  background: rgb(255 255 255 / 0.92);
}

.dark .top-music {
  color: rgb(255 255 255 / 0.9);
}

:global(.dark) .top-music__range::-webkit-slider-runnable-track {
  background: rgb(255 255 255 / 0.34);
}

:global(.dark) .top-music__range::-webkit-slider-thumb {
  background: rgb(255 255 255 / 0.88);
}

:global(.dark) .top-music__range::-moz-range-track {
  background: rgb(255 255 255 / 0.34);
}

:global(.dark) .top-music__range::-moz-range-thumb {
  background: rgb(255 255 255 / 0.88);
}

.top-music-track-enter-active,
.top-music-track-leave-active {
  transition: all 0.18s ease;
}

.top-music-track-enter-from,
.top-music-track-leave-to {
  opacity: 0;
  transform: translateX(-4px);
}
</style>
