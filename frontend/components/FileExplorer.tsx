'use client'
import React, { useEffect, useState } from 'react'
import api from '@/lib/api';

const FileExplorer = () => {
    const [data, setData] = useState('');

    useEffect(() => {
      const fetchData = async () => {
        try {
          const command = 'getfs';
          const response = await api.post('/execute', { command });
          console.log('Response:', response);
          setData(response.data);
        } catch (error) {
          console.error('Error fetching data:', error);
        }
      };
      fetchData();
    }, []);
  return (
    <div>FileExplorer</div>
  )
}

export default FileExplorer