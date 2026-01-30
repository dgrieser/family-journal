import { useTranslation } from 'react-i18next';

const LanguageSwitcher = () => {
  const { i18n } = useTranslation();
  return (
    <select
      className="border rounded px-2 py-1 text-sm"
      value={i18n.language}
      onChange={(event) => void i18n.changeLanguage(event.target.value)}
    >
      <option value="en">EN</option>
      <option value="de">DE</option>
    </select>
  );
};

export default LanguageSwitcher;
