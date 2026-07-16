<template>
  <div 
    :style="{ width: size + 'px', height: size + 'px' }" 
    class="relative select-none pointer-events-none mascot-container transition-transform duration-500 ease-spring"
    :class="[hovered ? 'scale-110' : 'scale-100']"
    @mouseenter="hovered = true"
    @mouseleave="hovered = false"
  >
    <svg 
      viewBox="0 0 100 100" 
      class="w-full h-full overflow-visible animate-breathe"
      :class="{ 'animate-jump-n-vibe': mood === 'excited', 'animate-thinking-sway': mood === 'thinking' }"
      :aria-label="'Mascote Divi ' + mood"
    >
      <!-- Body Shadow (subtle depth) -->
      <path
        :d="blobPath"
        fill="black"
        fill-opacity="0.05"
        transform="translate(2, 3)"
        class="transition-all duration-700 ease-spring"
      />
      <!-- Blob Body -->
      <path
        :d="blobPath"
        :fill="fillColor"
        class="transition-all duration-700 ease-spring"
      />
      
      <!-- Embellishments: Blush -->
      <g
        v-if="['happy', 'excited', 'proud', 'surprised'].includes(mood)"
        class="blush opacity-40"
      >
        <circle
          cx="30"
          cy="50"
          r="3"
          fill="white"
        />
        <circle
          cx="70"
          cy="50"
          r="3"
          fill="white"
        />
      </g>

      <!-- Eyes -->
      <g class="eyes animate-blink">
        <template v-if="['happy', 'proud'].includes(mood)">
          <circle
            cx="38"
            cy="42"
            r="3"
            fill="black"
          />
          <circle
            cx="62"
            cy="42"
            r="3"
            fill="black"
          />
          <circle
            v-if="mood === 'proud'"
            cx="39"
            cy="40.5"
            r="1.2"
            fill="white"
          />
          <circle
            v-if="mood === 'proud'"
            cx="63"
            cy="40.5"
            r="1.2"
            fill="white"
          />
        </template>
        <template v-else-if="mood === 'chill'">
          <line
            x1="34"
            y1="44"
            x2="42"
            y2="44"
            stroke="black"
            stroke-width="2.5"
            stroke-linecap="round"
          />
          <line
            x1="58"
            y1="44"
            x2="66"
            y2="44"
            stroke="black"
            stroke-width="2.5"
            stroke-linecap="round"
          />
        </template>
        <template v-else-if="mood === 'surprised'">
          <circle
            cx="38"
            cy="42"
            r="4"
            fill="white"
            stroke="black"
            stroke-width="1.5"
          />
          <circle
            cx="62"
            cy="42"
            r="4"
            fill="white"
            stroke="black"
            stroke-width="1.5"
          />
          <circle
            cx="38"
            cy="42"
            r="1.5"
            fill="black"
          />
          <circle
            cx="62"
            cy="42"
            r="1.5"
            fill="black"
          />
        </template>
        <template v-else-if="mood === 'excited'">
          <circle
            cx="38"
            cy="40"
            r="4.5"
            fill="black"
          />
          <circle
            cx="62"
            cy="40"
            r="4.5"
            fill="black"
          />
          <circle
            cx="40"
            cy="38"
            r="1.5"
            fill="white"
          />
          <circle
            cx="64"
            cy="38"
            r="1.5"
            fill="white"
          />
        </template>
        <template v-else-if="mood === 'thinking'">
          <circle
            cx="38"
            cy="40"
            r="3"
            fill="black"
          />
          <path
            d="M58 44 Q62 38 66 44"
            stroke="black"
            stroke-width="2.5"
            fill="none"
            stroke-linecap="round"
          />
        </template>
        <template v-else-if="mood === 'sleeping'">
          <path
            d="M34 44 Q38 48 42 44"
            stroke="black"
            stroke-width="2.5"
            fill="none"
            stroke-linecap="round"
          />
          <path
            d="M58 44 Q62 48 66 44"
            stroke="black"
            stroke-width="2.5"
            fill="none"
            stroke-linecap="round"
          />
        </template>
        <template v-else-if="mood === 'sad'">
          <circle
            cx="38"
            cy="44"
            r="3"
            fill="black"
          />
          <circle
            cx="62"
            cy="44"
            r="3"
            fill="black"
          />
          <line
            x1="34"
            y1="42"
            x2="40"
            y2="44"
            stroke="white"
            stroke-width="1.2"
            stroke-linecap="round"
          />
          <line
            x1="60"
            y1="44"
            x2="66"
            y2="42"
            stroke="white"
            stroke-width="1.2"
            stroke-linecap="round"
          />
        </template>
      </g>

      <!-- Mouth -->
      <g class="mouth">
        <path
          v-if="mood === 'happy' || mood === 'proud'"
          d="M42 62 Q50 70 58 62"
          stroke="black"
          stroke-width="2.5"
          fill="none"
          stroke-linecap="round"
        />
        <line
          v-else-if="mood === 'chill' || mood === 'sleeping'"
          x1="45"
          y1="64"
          x2="55"
          y2="64"
          stroke="black"
          stroke-width="2"
          stroke-linecap="round"
        />
        <circle
          v-else-if="mood === 'surprised'"
          cx="50"
          cy="65"
          r="3"
          fill="black"
        />
        <path
          v-else-if="mood === 'excited'"
          d="M40 60 Q50 72 60 60 L40 60"
          fill="black"
        />
        <path
          v-else-if="mood === 'thinking'"
          d="M45 65 Q50 62 55 65"
          stroke="black"
          stroke-width="2"
          fill="none"
          stroke-linecap="round"
        />
        <path
          v-else-if="mood === 'sad'"
          d="M42 68 Q50 62 58 68"
          stroke="black"
          stroke-width="2.5"
          fill="none"
          stroke-linecap="round"
        />
      </g>

      <!-- Stick limbs (Noodle style) -->
      <g
        class="limbs"
        stroke="black"
        stroke-width="3.2"
        stroke-linecap="round"
        fill="none"
        opacity="0.8"
      >
        <!-- Legs -->
        <path
          d="M38 80 Q35 88 32 92"
          class="animate-leg-left"
        />
        <path
          d="M62 80 Q65 88 68 92"
          class="animate-leg-right"
        />
        <!-- Arms -->
        <template v-if="mood === 'happy' || mood === 'excited'">
          <path
            d="M22 55 Q12 52 8 40"
            :class="{ 'animate-arm-wave': mood === 'happy', 'animate-excited-arms': mood === 'excited' }"
          />
          <path
            d="M78 55 Q88 52 92 40"
            :class="{ 'animate-excited-arms': mood === 'excited' }"
          />
        </template>
        <template v-else-if="mood === 'proud'">
          <path d="M22 62 Q15 68 12 78" />
          <path d="M78 62 Q85 68 88 78" />
        </template>
        <template v-else-if="mood === 'thinking'">
          <path d="M22 60 Q15 65 12 75" />
          <path d="M78 60 Q65 65 52 68" />
        </template>
        <template v-else-if="mood === 'sleeping' || mood === 'chill'">
          <path d="M25 65 Q18 72 15 80" />
          <path d="M75 65 Q82 72 85 80" />
        </template>
        <template v-else-if="mood === 'sad'">
          <path d="M22 62 Q18 72 16 82" />
          <path d="M78 62 Q82 72 84 82" />
        </template>
        <template v-else>
          <path d="M22 60 Q15 68 12 75" />
          <path d="M78 60 Q85 68 88 75" />
        </template>
      </g>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';

const props = defineProps({
  variant: {
    type: String,
    default: 'ember',
    validator: (v: string) => ['ember', 'meadow', 'sky', 'sunburst', 'flamingo', 'coral'].includes(v)
  },
  size: {
    type: Number,
    default: 100
  },
  mood: {
    type: String,
    default: 'happy',
    validator: (v: string) => ['happy', 'chill', 'surprised', 'excited', 'thinking', 'sleeping', 'proud', 'sad'].includes(v)
  }
});

const hovered = ref(false);

const fillColor = computed(() => {
  const colors: Record<string, string> = {
    ember: '#ff3e00',
    meadow: '#00a83d',
    sky: '#0090ff',
    sunburst: '#ffbb26',
    flamingo: '#ff58ae',
    coral: '#ff6b6b'
  };
  return colors[props.variant] || colors.ember;
});

const blobPath = computed(() => {
  // Base path: "M20,50 Q20,15 50,20 Q80,25 85,55 Q90,85 50,80 Q10,75 15,50 Z"
  const paths: Record<string, string> = {
    happy: "M20,50 Q20,15 50,20 Q80,25 85,55 Q90,85 50,80 Q10,75 15,50 Z",
    excited: "M20,48 Q20,13 50,18 Q80,23 85,53 Q90,83 50,78 Q10,73 15,48 Z",
    thinking: "M22,50 Q22,15 52,20 Q82,25 87,55 Q92,85 52,80 Q12,75 17,50 Z",
    sleeping: "M20,52 Q20,17 50,22 Q80,27 85,57 Q90,87 50,82 Q10,77 15,52 Z",
    chill: "M20,50 Q20,15 50,20 Q80,25 85,55 Q90,85 50,80 Q10,75 15,50 Z",
    proud: "M20,49 Q20,14 50,19 Q80,24 85,54 Q90,84 50,79 Q10,74 15,49 Z",
    surprised: "M20,47 Q20,12 50,17 Q80,22 85,52 Q90,82 50,77 Q10,72 15,47 Z",
    sad: "M20,52 Q20,25 50,30 Q80,35 85,60 Q90,85 50,80 Q10,75 15,52 Z"
  };
  return paths[props.mood] || paths.happy;
});
</script>

<style scoped>
@keyframes breathe {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.015, 0.985); }
}

@keyframes blink {
  0%, 90%, 100% { transform: scaleY(1); }
  95% { transform: scaleY(0.1); }
}

@keyframes jump-n-vibe {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-5px) rotate(2deg); }
}

@keyframes arm-wave {
  0%, 100% { transform: rotate(0deg); }
  50% { transform: rotate(-15deg); }
}

@keyframes excited-arms {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-3px) rotate(5deg); }
}

@keyframes thinking-sway {
  0%, 100% { transform: rotate(-1deg); }
  50% { transform: rotate(1deg); }
}

@keyframes leg-wiggle {
  0%, 100% { transform: rotate(0deg); }
  50% { transform: rotate(3deg); }
}

.animate-breathe {
  animation: breathe 5s ease-in-out infinite;
  transform-origin: center bottom;
}

.animate-blink {
  animation: blink 6s infinite;
  transform-origin: center 42px;
}

.animate-jump-n-vibe {
  animation: jump-n-vibe 0.4s ease-in-out infinite;
}

.animate-arm-wave {
  animation: arm-wave 2s ease-in-out infinite;
  transform-origin: 22px 55px;
}

.animate-excited-arms {
  animation: excited-arms 0.2s ease-in-out infinite;
  transform-origin: center;
}

.animate-thinking-sway {
  animation: thinking-sway 4s ease-in-out infinite;
  transform-origin: center bottom;
}

.animate-leg-left {
  animation: leg-wiggle 3s ease-in-out infinite;
  transform-origin: 35px 78px;
}

.animate-leg-right {
  animation: leg-wiggle 3s ease-in-out infinite reverse;
  transform-origin: 65px 78px;
}

.mascot-container:hover .eyes {
  transform: translateY(-1px);
  transition: transform 0.3s var(--ease-spring);
}
</style>

