'use client';

import { useFormContext } from 'react-hook-form';
import { Target, AlertTriangle } from 'lucide-react';

interface MatchingSectionProps {
    missingFields?: string[];
}

export default function MatchingSection({ missingFields = [] }: MatchingSectionProps) {
    const { register } = useFormContext();

    const getInputClass = (fieldName: string) => {
        const baseClass = "w-full px-3 py-2 border rounded-lg focus:ring-2 text-slate-900 placeholder-slate-400 bg-white transition-colors";
        const isMissing = missingFields.includes(fieldName);

        if (isMissing) {
            return `${baseClass} border-red-300 bg-red-50 focus:ring-red-500 focus:border-red-500`;
        }

        return `${baseClass} border-slate-300 focus:ring-blue-500 focus:border-blue-500`;
    };

    return (
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
            <div className="mb-6">
                <div className="flex items-center space-x-3 mb-2">
                    <div className="w-8 h-8 bg-emerald-100 rounded-lg flex items-center justify-center">
                        <Target className="h-4 w-4 text-emerald-600" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900">Advanced Matching Configuration</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Advanced security parameters for privacy-preserving matching. Patient records are always matched using both bloom filters and min hashes.
                </p>
            </div>

            {/* Warning Box */}
            <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-6">
                <div className="flex items-start space-x-3">
                    <AlertTriangle className="h-5 w-5 text-amber-600 mt-0.5" />
                    <div>
                        <h4 className="text-sm font-medium text-amber-900 mb-1">Important: Coordination Required</h4>
                        <p className="text-sm text-amber-700">
                            These values must be <strong>identical across both parties</strong>. Different values will result in unpredictable matches.
                            Only modify these settings if you need enhanced security and have coordinated with the other party.
                        </p>
                    </div>
                </div>
            </div>

            <div className="space-y-6">
                {/* Bloom Filter Settings */}
                <div>
                    <h4 className="text-md font-medium text-slate-800 mb-4">Bloom Filter Parameters</h4>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Bloom Filter Size (bits)
                            </label>
                            <input
                                type="number"
                                {...register('matching.bloom_filter_size', { valueAsNumber: true })}
                                placeholder="1024"
                                min="256"
                                max="8192"
                                className={getInputClass('matching.bloom_filter_size')}
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                Larger filters improve accuracy but use more memory. Default: 1024 bits.
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Number of Hash Functions
                            </label>
                            <input
                                type="number"
                                {...register('matching.num_hash_functions', { valueAsNumber: true })}
                                placeholder="4"
                                min="1"
                                max="20"
                                className={getInputClass('matching.num_hash_functions')}
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                More hash functions increase security but may reduce accuracy. Default: 4.
                            </p>
                        </div>
                    </div>
                </div>

                {/* MinHash Settings */}
                <div>
                    <h4 className="text-md font-medium text-slate-800 mb-4">MinHash Parameters</h4>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                MinHash Signature Size
                            </label>
                            <input
                                type="number"
                                {...register('matching.minhash_size', { valueAsNumber: true })}
                                placeholder="128"
                                min="64"
                                max="512"
                                className={getInputClass('matching.minhash_size')}
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                Larger signatures improve similarity detection. Default: 128.
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Q-gram Size
                            </label>
                            <input
                                type="number"
                                {...register('matching.qgram_size', { valueAsNumber: true })}
                                placeholder="2"
                                min="1"
                                max="5"
                                className={getInputClass('matching.qgram_size')}
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                Character n-gram size for string comparison. Default: 2 (bigrams).
                            </p>
                        </div>
                    </div>
                </div>

                {/* Threshold Settings */}
                <div>
                    <h4 className="text-md font-medium text-slate-800 mb-4">Matching Thresholds</h4>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Similarity Threshold (0.0 - 1.0)
                            </label>
                            <input
                                type="number"
                                step="0.01"
                                min="0"
                                max="1"
                                {...register('matching.threshold', { valueAsNumber: true })}
                                placeholder="0.7"
                                className={getInputClass('matching.threshold')}
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                Minimum similarity score to consider records a match. Default: 0.7.
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Noise Level (0.0 - 0.1)
                            </label>
                            <input
                                type="number"
                                step="0.01"
                                min="0"
                                max="0.1"
                                {...register('matching.noise_level', { valueAsNumber: true })}
                                placeholder="0.01"
                                className={getInputClass('matching.noise_level')}
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                Adds privacy protection. Higher values = more private, less accurate. Default: 0.01.
                            </p>
                        </div>
                    </div>
                </div>

                {/* Performance Settings */}
                <div>
                    <h4 className="text-md font-medium text-slate-800 mb-4">Performance Optimization</h4>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label className="flex items-center space-x-3">
                                <input
                                    type="checkbox"
                                    {...register('matching.blocking_enabled')}
                                    className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                                />
                                <span className="text-sm font-medium text-slate-700">Enable Blocking</span>
                            </label>
                            <p className="mt-1 text-xs text-slate-600 ml-6">
                                Group similar records to improve performance on large datasets.
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Block Size
                            </label>
                            <input
                                type="number"
                                {...register('matching.block_size', { valueAsNumber: true })}
                                placeholder="1000"
                                min="100"
                                max="10000"
                                className={getInputClass('matching.block_size')}
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                Maximum records per block. Default: 1000.
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
} 