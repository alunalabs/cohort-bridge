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
        matching: {
            bloom_size: 1024,
            bloom_hashes: 4,
            minhash_size: 128,
            qgram_length: 2,
            hamming_threshold: 100,
            jaccard_threshold: 0.6,
            qgram_threshold: 0.7,
            noise_level: 0.01,
        },
    };

    return (
        <FlexibleConfigBuilder
            title="Basic Configuration"
            description="Simple setup for basic privacy-preserving record linkage"
            defaultSections={['database', 'peer', 'matching']}
            defaultValues={defaultValues}
            icon={Settings}
        />
    );
} 