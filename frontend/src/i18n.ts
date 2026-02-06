import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

const resources = {
  en: {
    translation: {
      "app_name": "FamilyJournal",
      "login": "Login",
      "register": "Register",
      "email": "Email",
      "password": "Password",
      "logout": "Logout",
      "timeline": "Timeline",
      "persons": "Persons",
      "admin": "Admin",
      "save": "Save",
      "cancel": "Cancel",
      "delete": "Delete",
      "edit": "Edit",
      "new_post": "New Post",
      "add_comment": "Add Comment",
      "hashtags": "Hashtags",
      "mentions": "Mentions",
      "date": "Date",
      "search": "Search",
      "no_posts": "No posts for this day.",
      "role": "Role",
      "user": "User",
      "admin_role": "Admin",
      "name": "Name",
      "description": "Description",
      "attachments": "Attachments",
      "upload_files": "Upload Files",
      "loading": "Loading...",
      "promote": "Promote",
      "demote": "Demote",
      "active": "Active",
      "inactive": "Inactive",
      "profile": "Profile",
      "update": "Update",
      "leave_blank": "leave blank to leave unchanged",
      "success": "Success",
      "error": "Error",
      "activate": "Activate",
      "deactivate": "Deactivate",
      "clear_filters": "Clear all filters",
      "dont_have_account": "Don't have an account?",
      "already_have_account": "Already have an account?",
      "registration_success_login": "Registration successful. Please log in.",
    }
  },
  de: {
    translation: {
      "app_name": "FamilienJournal",
      "login": "Anmelden",
      "register": "Registrieren",
      "email": "E-Mail",
      "password": "Passwort",
      "logout": "Abmelden",
      "timeline": "Timeline",
      "persons": "Personen",
      "admin": "Admin",
      "save": "Speichern",
      "cancel": "Abbrechen",
      "delete": "Löschen",
      "edit": "Bearbeiten",
      "new_post": "Neuer Eintrag",
      "add_comment": "Kommentar hinzufügen",
      "hashtags": "Hashtags",
      "mentions": "Erwähnungen",
      "date": "Datum",
      "search": "Suche",
      "no_posts": "Keine Einträge für diesen Tag.",
      "role": "Rolle",
      "user": "Benutzer",
      "admin_role": "Admin",
      "name": "Name",
      "description": "Beschreibung",
      "attachments": "Anhänge",
      "upload_files": "Dateien hochladen",
      "loading": "Laden...",
      "promote": "Befördern",
      "demote": "Herabstufen",
      "active": "Aktiv",
      "inactive": "Inaktiv",
      "profile": "Profil",
      "update": "Aktualisieren",
      "leave_blank": "leer lassen um nicht zu ändern",
      "success": "Erfolg",
      "error": "Fehler",
      "activate": "Aktivieren",
      "deactivate": "Deaktivieren",
      "clear_filters": "Alle Filter löschen",
      "dont_have_account": "Noch kein Konto?",
      "already_have_account": "Schon ein Konto?",
      "registration_success_login": "Registrierung erfolgreich. Bitte anmelden.",
    }
  }
};

i18n
  .use(initReactI18next)
  .init({
    resources,
    lng: "de",
    fallbackLng: "en",
    interpolation: {
      escapeValue: false
    }
  });

export default i18n;
