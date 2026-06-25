<template>
  <div
    v-if="$route.path !== '/new' && $route.path.indexOf('/edit/') < 0"
    class="header relative mb-14"
  >
    <div
      v-if="$route.path !== '/' && $route.path.indexOf('/memo/') < 0"
      :class="{ 'bg-[#4c4c4c]/80 z-10': y > 100 }"
      class="flex fixed justify-between items-center p-4 w-full md:w-[567px] text-white top-0"
    >
      <NuxtLink class="flex items-center" title="返回主页">
        <UIcon
          @click="navigateTo('/')"
          name="i-carbon-chevron-left"
          class="w-5 h-5 cursor-pointer mr-4"
        />
        <span v-if="$route.path === '/user/calendar'">日历检索</span>
        <span v-else-if="$route.path === '/sys/settings'">系统设置</span>
        <span v-else-if="$route.path === '/user/settings'">用户中心</span>
        <span v-else-if="$route.path.indexOf('/tags/') >= 0">
          {{ route.params.tag || "话题专栏" }}
        </span>
        <span v-else-if="$route.path === '/friend'">友情链接</span>
        <span v-else>
          <span v-if="!global.userinfo.token && $route.path === '/user/login'">
            登录
          </span>
          <span
            v-else-if="!global.userinfo.token && $route.path === '/user/reg'"
          >
            注册
          </span>
          <span v-else>{{ props.user.nickname }} 的空间</span>
        </span>
      </NuxtLink>
      <NuxtLink
        v-if="$route.path === '/user/settings' && global.userinfo.token"
        class="hidden sm:flex"
        title="登出"
        @click="logout"
      >
        <UIcon name="i-carbon-logout" class="w-5 h-5 cursor-pointer" />
      </NuxtLink>
      <span
        v-if="$route.path === '/friend' && global.userinfo.id === 1"
        class="flex"
      >
        <UIcon
          name="i-carbon-add"
          class="w-6 h-6 cursor-pointer"
          @click="$emit('add-friend')"
        />
      </span>
    </div>

    <div
      v-if="$route.path === '/' && sysConfig.music?.url"
      class="absolute top-3 left-3 z-20 hidden sm:block"
    >
      <TopMusicPlayer :music="sysConfig.music" />
    </div>

    <div
      class="dark:bg-neutral-800 hidden sm:flex sm:absolute sm:-right-10 sm:rounded sm:p-2 sm:flex-col sm:w-fit justify-end shadow w-full flex-row top-0 p-1 flex gap-2 bg-white"
    >
      <svg
        v-if="mode.value === 'light'"
        class="lucide lucide-moon-star-icon cursor-pointer"
        @click="toggleMode"
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="#FDE047"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9"></path>
        <path d="M20 3v4"></path>
        <path d="M22 5h-4"></path>
      </svg>

      <svg
        v-else
        class="lucide lucide-sun-icon cursor-pointer"
        @click="toggleMode"
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="#FDE047"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <circle cx="12" cy="12" r="4"></circle>
        <path d="M12 2v2"></path>
        <path d="M12 20v2"></path>
        <path d="m4.93 4.93 1.41 1.41"></path>
        <path d="m17.66 17.66 1.41 1.41"></path>
        <path d="M2 12h2"></path>
        <path d="M20 12h2"></path>
        <path d="m6.34 17.66-1.41 1.41"></path>
        <path d="m19.07 4.93-1.41 1.41"></path>
      </svg>

      <NuxtLink v-if="global.userinfo.token" to="/new" title="发表">
        <UIcon
          name="i-carbon-camera"
          class="text-[#9fc84a] w-5 h-5 cursor-pointer"
        />
      </NuxtLink>
      <NuxtLink
        v-if="$route.path !== '/user/calendar' && global.userinfo.token"
        to="/user/calendar"
        title="日历检索"
      >
        <UIcon
          name="i-jam-search-folder"
          class="text-[#9fc84a] w-5 h-5 cursor-pointer"
        />
      </NuxtLink>
      <NuxtLink
        v-if="$route.path !== '/sys/settings' && global.userinfo.id === 1"
        to="/sys/settings"
        title="系统设置"
      >
        <UIcon
          name="i-carbon-settings"
          class="text-[#9fc84a] w-5 h-5 cursor-pointer"
        />
      </NuxtLink>
      <NuxtLink
        v-if="$route.path !== '/user/settings' && global.userinfo.token"
        to="/user/settings"
        title="用户中心"
      >
        <UIcon
          name="i-carbon-user-avatar"
          class="text-[#9fc84a] w-5 h-5 cursor-pointer"
        />
      </NuxtLink>
      <NuxtLink v-if="!global.userinfo.token" to="/user/login" title="登录">
        <UIcon
          name="i-carbon-login"
          class="text-[#9fc84a] w-5 h-5 cursor-pointer"
        />
      </NuxtLink>
    </div>

    <div
      v-if="$route.path === '/' && sysConfig.music?.url"
      class="absolute top-3 left-3 z-20 sm:hidden"
    >
      <TopMusicPlayer :music="sysConfig.music" />
    </div>

    <img class="header-img w-full" :src="props.user.coverUrl" alt="" />
    <div
      v-if="showWeatherWidget"
      class="absolute left-4 bottom-4 z-10 max-w-[calc(100%-120px)] overflow-hidden rounded-2xl border border-white/12 bg-black/22 px-3.5 py-2.5 text-white shadow-[0_8px_24px_rgba(0,0,0,0.18)] backdrop-blur-md transition-all dark:border-white/10 dark:bg-black/28 dark:shadow-[0_10px_28px_rgba(0,0,0,0.3)] sm:left-6 sm:bottom-6"
    >
      <div class="absolute inset-0 bg-gradient-to-br from-white/14 via-white/6 to-transparent dark:from-white/10 dark:via-white/4"></div>
      <div class="relative flex items-center gap-2.5">
        <span class="h-1.5 w-1.5 flex-shrink-0 rounded-full bg-white/40 shadow-[0_0_0_3px_rgba(255,255,255,0.08)] dark:bg-white/35 dark:shadow-[0_0_0_3px_rgba(255,255,255,0.05)]"></span>
        <div class="min-w-0 weather-text flex items-center gap-2 text-sm leading-none text-white/95">
          <span class="truncate max-w-[6rem]">{{ weatherLocationText }}</span>
          <span aria-hidden="true">{{ weatherEmoji }}</span>
          <span class="truncate max-w-[4rem]">{{ weatherStatusText }}</span>
          <span>{{ weatherTempText }}</span>
          <span class="weather-badge">{{ weatherAqiText }}</span>
          <span class="truncate max-w-[5rem]">{{ weatherWindText }}</span>
        </div>
      </div>
    </div>
    <div class="absolute right-2 bottom-[-40px]">
      <div class="userinfo flex flex-col">
        <div class="flex flex-row items-center gap-4 justify-end">
          <div class="username text-lg font-bold text-white">
            {{ props.user.nickname }}
          </div>
          <img
            :src="props.user.avatarUrl"
            class="avatar w-[70px] h-[70px] rounded-xl"
          />
        </div>
        <div class="slogon text-gray truncate w-full text-end text-xs mt-2">
          {{ props.user.slogan }}
        </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { toast } from "vue-sonner";
import { computed, ref, watch } from "vue";
import type { SysConfigVO, UserVO } from "~/types";
import { useGlobalState } from "~/store";

const global = useGlobalState();
const route = useRoute();
const sysConfig = useState<SysConfigVO>("sysConfig");

const props = defineProps<{ user: UserVO }>();
const mode = useColorMode();
const { y } = useWindowScroll();
const showWeatherWidget = computed(
  () => route.path === "/" && Boolean(sysConfig.value.enableVisitorWeather),
);

type ClientWeatherResp = {
  code?: number;
  message?: string;
  data?: {
    location?: {
      name?: string;
      province?: string;
      city?: string;
      county?: string;
    };
    weather?: {
      condition?: string;
      temperature?: number;
      wind_direction?: string;
      wind_power?: string;
    };
    air_quality?: {
      aqi?: number;
      quality?: string;
    };
  };
};

type WeatherViewModel = {
  location: string;
  weather: string;
  temperature: string;
  airQuality: string;
  wind: string;
};

const weatherInfo = ref<WeatherViewModel | null>(null);
const weatherLoading = ref(false);
const weatherError = ref("");

const weatherEmoji = computed(() => {
  const weather = weatherInfo.value?.weather ?? "";
  if (weather.includes("雷")) return "⛈️";
  if (weather.includes("雪")) return "❄️";
  if (weather.includes("雨")) return "🌧️";
  if (weather.includes("雾")) return "🌫️";
  if (weather.includes("晴")) return "☀️";
  if (weather.includes("云") || weather.includes("阴")) return "⛅";
  return "🌤️";
});

const weatherLocationText = computed(() => {
  if (weatherLoading.value) return "正在获取天气...";
  if (weatherError.value) return "天气获取失败";
  return weatherInfo.value?.location || "天气获取失败";
});

const weatherStatusText = computed(() => {
  if (weatherLoading.value) return "请稍候";
  if (weatherError.value) return weatherError.value;
  return weatherInfo.value?.weather || "未知";
});

const weatherTempText = computed(() => {
  if (weatherLoading.value || weatherError.value) return "--";
  return weatherInfo.value?.temperature || "--";
});

const weatherAqiText = computed(() => {
  if (weatherLoading.value || weatherError.value) return "--";
  return weatherInfo.value?.airQuality || "--";
});

const weatherWindText = computed(() => {
  if (weatherLoading.value || weatherError.value) return "--";
  return weatherInfo.value?.wind || "--";
});

const logout = async () => {
  global.value.userinfo = {};
  await navigateTo("/");
};

const toggleMode = () => {
  if (mode.preference === "system") {
    mode.preference = "dark";
  } else if (mode.preference === "dark") {
    mode.preference = "light";
  } else {
    mode.preference = "system";
    toast.success("显示模式将跟随系统设置");
  }
};

const formatAQI = (value: string | number | null | undefined) => {
  if (value === null || value === undefined || value === "") return "--";
  return String(value);
};

const getWeatherQuery = (): string => {
  return sysConfig.value.visitorWeatherCity?.trim() || "北京";
};

const normalizeWeather = (weather: ClientWeatherResp): WeatherViewModel => {
  const location = (weather.data?.location?.city || weather.data?.location?.name || "未知位置").replace(/市$/, "");

  return {
    location,
    weather: weather.data?.weather?.condition || "未知",
    temperature: weather.data?.weather?.temperature !== undefined
      ? String(weather.data.weather.temperature) + "°C"
      : "--",
    airQuality: weather.data?.air_quality?.quality || formatAQI(weather.data?.air_quality?.aqi),
    wind: [weather.data?.weather?.wind_direction, weather.data?.weather?.wind_power]
      .filter(Boolean)
      .join(" ") || "--",
  };
};

const loadWeather = async () => {
  if (!showWeatherWidget.value) {
    weatherInfo.value = null;
    weatherLoading.value = false;
    weatherError.value = "";
    return;
  }

  weatherLoading.value = true;
  weatherError.value = "";

  try {
    const query = getWeatherQuery();
    const weather = await $fetch<ClientWeatherResp>("/api/weather", {
      query: { query },
    });

    if (weather.code && weather.code !== 200) {
      throw new Error(weather.message || "天气服务返回异常");
    }

    weatherInfo.value = normalizeWeather(weather);
  } catch (error) {
    weatherInfo.value = null;
    weatherError.value = error instanceof Error ? error.message : "天气获取失败";
  } finally {
    weatherLoading.value = false;
  }
};

watch(
  () => [showWeatherWidget.value, sysConfig.value.visitorWeatherCity] as const,
  async ([enabled], old) => {
    const prev = old ?? [undefined, undefined] as unknown as [boolean, string];
    const [prevEnabled, prevCity] = prev;
    if (!enabled) {
      weatherInfo.value = null;
      weatherError.value = "";
      weatherLoading.value = false;
      return;
    }
    if (enabled === prevEnabled && sysConfig.value.visitorWeatherCity === prevCity) {
      return;
    }
    await loadWeather();
  },
  { immediate: true },
);
</script>

<style scoped>
.weather-text {
  flex-wrap: nowrap;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.18);
}

.weather-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.6rem;
  height: 1.45rem;
  padding: 0 0.45rem;
  border-radius: 0.55rem;
  background: linear-gradient(180deg, #8df46a 0%, #63cf45 100%);
  color: #fff;
  font-size: 0.95rem;
  line-height: 1;
  box-shadow: 0 6px 16px rgba(99, 207, 69, 0.22);
}
</style>
