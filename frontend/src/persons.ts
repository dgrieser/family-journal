import api from './api';
import type { PaginatedResponse, Person } from './types';

const DEFAULT_PERSON_SEARCH_PAGE_SIZE = 10;

export const searchPersons = async (
  search: string,
  pageSize = DEFAULT_PERSON_SEARCH_PAGE_SIZE
): Promise<Person[]> => {
  const response = await api.get<PaginatedResponse<Person>>('/persons', {
    params: {
      page: 1,
      pageSize,
      search
    }
  });

  return response.data.items;
};
