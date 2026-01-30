import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { apiFetch } from '../api/client';

interface Person {
  id: number;
  name: string;
  description?: string;
}

const PersonsPage = () => {
  const { t } = useTranslation();
  const [persons, setPersons] = useState<Person[]>([]);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editingName, setEditingName] = useState('');
  const [editingDescription, setEditingDescription] = useState('');

  const loadPersons = async () => {
    const data = await apiFetch('/persons');
    setPersons(data);
  };

  useEffect(() => {
    void loadPersons();
  }, []);

  const handleAdd = async (event: React.FormEvent) => {
    event.preventDefault();
    await apiFetch('/persons', {
      method: 'POST',
      body: JSON.stringify({ name, description: description || null })
    });
    setName('');
    setDescription('');
    await loadPersons();
  };

  const handleDelete = async (id: number) => {
    await apiFetch(`/persons/${id}`, { method: 'DELETE' });
    await loadPersons();
  };

  const startEdit = (person: Person) => {
    setEditingId(person.id);
    setEditingName(person.name);
    setEditingDescription(person.description || '');
  };

  const handleUpdate = async () => {
    if (editingId === null) return;
    await apiFetch(`/persons/${editingId}`, {
      method: 'PUT',
      body: JSON.stringify({ name: editingName, description: editingDescription || null })
    });
    setEditingId(null);
    setEditingName('');
    setEditingDescription('');
    await loadPersons();
  };

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-4">
      <h2 className="text-2xl font-semibold">{t('persons.title')}</h2>
      <form onSubmit={handleAdd} className="bg-white p-4 rounded shadow-sm space-y-3">
        <input
          className="w-full border rounded px-3 py-2"
          placeholder={t('persons.name')}
          value={name}
          onChange={(event) => setName(event.target.value)}
        />
        <input
          className="w-full border rounded px-3 py-2"
          placeholder={t('persons.description')}
          value={description}
          onChange={(event) => setDescription(event.target.value)}
        />
        <button className="bg-slate-900 text-white px-3 py-2 rounded" type="submit">
          {t('persons.add')}
        </button>
      </form>
      <div className="space-y-2">
        {persons.map((person) => (
          <div key={person.id} className="bg-white p-3 rounded shadow-sm flex justify-between">
            <div>
              {editingId === person.id ? (
                <div className="space-y-2">
                  <input
                    className="w-full border rounded px-2 py-1"
                    value={editingName}
                    onChange={(event) => setEditingName(event.target.value)}
                  />
                  <input
                    className="w-full border rounded px-2 py-1"
                    value={editingDescription}
                    onChange={(event) => setEditingDescription(event.target.value)}
                  />
                </div>
              ) : (
                <>
                  <p className="font-semibold">{person.name}</p>
                  <p className="text-sm text-slate-500">{person.description}</p>
                </>
              )}
            </div>
            <div className="flex gap-2">
              {editingId === person.id ? (
                <button className="text-sm text-slate-600" onClick={() => void handleUpdate()}>
                  {t('persons.save')}
                </button>
              ) : (
                <button className="text-sm text-slate-600" onClick={() => startEdit(person)}>
                  {t('persons.edit')}
                </button>
              )}
              <button className="text-sm text-red-600" onClick={() => void handleDelete(person.id)}>
                {t('persons.delete')}
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default PersonsPage;
