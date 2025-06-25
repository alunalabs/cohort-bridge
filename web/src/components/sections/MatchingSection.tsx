'use client';

import { useFormContext } from 'react-hook-form';
import { Target } from 'lucide-react';

export default function MatchingSection() {
    const { register } = useFormContext();

    return (
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
            <div className="mb-6">
                <div className="flex items-center space-x-3 mb-2">
                    <div className="w-8 h-8 bg-indigo-100 rounded-lg flex items-center justify-center">
                        <Target className="h-4 w-4 text-indigo-600" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900">Matching Configuration</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Configure privacy-preserving matching algorithms. Balance privacy protection with matching accuracy.
                </p>
            </div>

            <div className="space-y-6">
                {/* Bloom Filter Settings */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                            Bloom Filter Size
                        </label>
                        <input
                            type="number"
                            {...register('matching.bloom_size', { valueAsNumber: true })}
                            placeholder="2048"
                            min="1"
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <p className="mt-1 text-xs text-slate-600">
                            Larger values (2048+) improve accuracy but use more memory. Start with 1024 for testing.
                        </p>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                            Bloom Filter Hash Functions
                        </label>
                        <input
                            type="number"
                            {...register('matching.bloom_hashes', { valueAsNumber: true })}
                            placeholder="8"
                            min="1"
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <p className="mt-1 text-xs text-slate-600">
                            More hash functions (4-12) reduce false matches but slow processing. 8 is optimal for most cases.
                        </p>
                    </div>
                </div>

                {/* MinHash Settings */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                            MinHash Signature Size
                        </label>
                        <input
                            type="number"
                            {...register('matching.minhash_size', { valueAsNumber: true })}
                            placeholder="256"
                            min="1"
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <p className="mt-1 text-xs text-slate-600">
                            Larger signatures (256+) improve similarity detection. 128 is good for most datasets.
                        </p>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                            Q-gram Length
                        </label>
                        <input
                            type="number"
                            {...register('matching.qgram_length', { valueAsNumber: true })}
                            placeholder="3"
                            min="1"
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <p className="mt-1 text-xs text-slate-600">
                            Character groups for fuzzy matching. 2=bigrams (fast), 3=trigrams (balanced), 4+ (precise).
                        </p>
                    </div>
                </div>

                {/* Threshold Settings */}
                <div className="space-y-4">
                    <h4 className="text-md font-medium text-slate-800">Matching Thresholds</h4>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Hamming Threshold
                            </label>
                            <input
                                type="number"
                                {...register('matching.hamming_threshold', { valueAsNumber: true })}
                                placeholder="200"
                                min="0"
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                Lower values = more matches but less precision. Adjust based on Bloom filter size.
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Jaccard Threshold
                            </label>
                            <input
                                type="number"
                                step="0.01"
                                min="0"
                                max="1"
                                {...register('matching.jaccard_threshold', { valueAsNumber: true })}
                                placeholder="0.75"
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                Set similarity requirement (0.6=loose, 0.8=strict). Higher = fewer but better matches.
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Q-gram Threshold
                            </label>
                            <input
                                type="number"
                                step="0.01"
                                min="0"
                                max="1"
                                {...register('matching.qgram_threshold', { valueAsNumber: true })}
                                placeholder="0.85"
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                String similarity cutoff. 0.7=tolerant of typos, 0.9=requires near-exact matches.
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Noise Level
                            </label>
                            <input
                                type="number"
                                step="0.01"
                                min="0"
                                max="1"
                                {...register('matching.noise_level', { valueAsNumber: true })}
                                placeholder="0.02"
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            <p className="mt-1 text-xs text-slate-600">
                                Adds privacy protection (0.01-0.05 recommended). Higher = more private, less accurate.
                            </p>
                        </div>
                    </div>
                </div>

                {/* Information Box */}
                <div className="bg-indigo-50 border border-indigo-200 rounded-lg p-4">
                    <div className="flex items-start space-x-3">
                        <div className="w-5 h-5 bg-indigo-100 rounded-full flex items-center justify-center mt-0.5">
                            <div className="w-2 h-2 bg-indigo-600 rounded-full"></div>
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-indigo-900 mb-1">Matching Algorithm</h4>
                            <p className="text-sm text-indigo-700">
                                These settings control the privacy-preserving record linkage algorithm.
                                Higher thresholds mean stricter matching, while larger filter sizes improve accuracy at the cost of memory usage.
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
} 