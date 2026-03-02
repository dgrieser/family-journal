import api from './api';
import type { PaginatedResponse, Person } from './types';

const PERSONS_PAGE_SIZE = 100;

export const fetchAllPersons = async (): Promise<Person[]> => {
  const firstPage = await api.get<PaginatedResponse<Person>>('/persons', {
    params: { page: 1, pageSize: PERSONS_PAGE_SIZE }
  });

  const persons = [...firstPage.data.items];
  const { totalPages } = firstPage.data.pagination;

  for (let page = 2; page <= totalPages; page += 1) {
    const response = await api.get<PaginatedResponse<Person>>('/persons', {
      params: { page, pageSize: PERSONS_PAGE_SIZE }
    });
    persons.push(...response.data.items);
  }

  return persons;
};
