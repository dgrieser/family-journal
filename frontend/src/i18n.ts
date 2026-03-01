import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

const resourcesBackend = {
  type: 'backend' as const,
  read(language: string, namespace: string, callback: (error: Error | null, data: unknown) => void) {
    if (namespace !== 'translation') {
      callback(null, {});
      return;
    }

    import(`./locales/${language}.json`)
      .then((module) => {
        callback(null, module.default);
      })
      .catch((error: unknown) => {
        callback(error instanceof Error ? error : new Error('Failed to load locale'), null);
      });
  }
};

i18n
  .use(resourcesBackend)
  .use(initReactI18next)
  .init({
    lng: 'de',
    fallbackLng: 'en',
    defaultNS: 'translation',
    ns: ['translation'],
    interpolation: {
      escapeValue: false
    }
  });

export default i18n;
