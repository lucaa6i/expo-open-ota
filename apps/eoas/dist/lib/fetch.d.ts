import { RequestInit, Response } from 'node-fetch';
export declare function fetchWithRetries(url: string, options: RequestInit): Promise<Response>;
