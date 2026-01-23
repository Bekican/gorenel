import axios from 'axios'
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:9090'

const apiClient = axios.create({
    baseURL : API_BASE_URL,
    timeout : 10000,
    headers : {
        'Content-Type' : 'application/json',
    },
});

//typelar