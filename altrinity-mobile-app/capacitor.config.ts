import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'com.altrinity.app',
  appName: 'Altrinity',
  webDir: 'dist',
  bundledWebRuntime: false,
  server: {
    iosScheme: 'capacitor',   // ✅ default
  },
};

export default config;
