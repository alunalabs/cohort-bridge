'use client';

import { Settings } from 'lucide-react';
import FlexibleConfigBuilder from '@/components/FlexibleConfigBuilder';

export default function BasicConfigPage() {
    const defaultValues = {
        database: {
            type: 'csv',
            filename: 'data/patients.csv',
            fields: ['name:first_name', 'name:last_name', 'date:date_of_birth', 'gender:gender', 'zip:zip_code'],
            random_bits_percent: 0.0,
        },
        peer: {
            host: 'localhost',
            port: 8080,
        },
        listen_port: 8080,
        security: {
            rate_limit_per_min: 5,
        },
        matching: {
            hamming_threshold: 20,
            jaccard_threshold: 0.32,
        },
    };

    return (
        <FlexibleConfigBuilder
            title="Basic Configuration"
            description="Simple setup for basic privacy-preserving record linkage with data normalization"
            defaultSections={['database', 'peer']}
            defaultValues={defaultValues}
            icon={Settings}
        />
    );
} 