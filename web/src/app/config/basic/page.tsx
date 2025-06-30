'use client';

import { Settings } from 'lucide-react';
import FlexibleConfigBuilder from '@/components/FlexibleConfigBuilder';

export default function BasicConfigPage() {
    const defaultValues = {
        database: {
            type: 'csv',
            filename: 'data/patients.csv',
            fields: ['first_name', 'last_name', 'date_of_birth', 'zip_code'],
            random_bits_percent: 0.0,
        },
        peer: {
            host: 'localhost',
            port: 8081,
        },
        listen_port: 8080,
        private_key: '',
    };

    return (
        <FlexibleConfigBuilder
            title="Basic Configuration"
            description="Simple setup for basic privacy-preserving record linkage"
            defaultSections={['database', 'peer']}
            defaultValues={defaultValues}
            icon={Settings}
        />
    );
} 