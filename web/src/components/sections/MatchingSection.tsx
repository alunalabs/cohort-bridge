'use client';

import { useFormContext } from 'react-hook-form';
import { Target, AlertTriangle, Info } from 'lucide-react';

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
                    <h3 className="text-lg font-semibold text-slate-900">Matching Thresholds</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Configure similarity thresholds for privacy-preserving record matching.
                </p>
            </div>

            {/* Warning Box */}
            <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-6">
                <div className="flex items-start space-x-3">
                    <AlertTriangle className="h-5 w-5 text-amber-600 mt-0.5" />
                    <div>
                        <h4 className="text-sm font-medium text-amber-900 mb-1">Important: Coordination Required</h4>
                        <p className="text-sm text-amber-700">
                            These threshold values must be <strong>identical across both parties</strong>. Different values will result in no matches.
                            Only modify these if you need different sensitivity levels and have coordinated with the other party.
                        </p>
                    </div>
                </div>
            </div>

            <div className="space-y-6">
                {/* Threshold Settings */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-2">
                            Hamming Threshold
                        </label>
                        <input
                            type="number"
                            {...register('matching.hamming_threshold', { valueAsNumber: true })}
                            placeholder="20"
                            min="1"
                            max="1000"
                            className={getInputClass('matching.hamming_threshold')}
                        />
                        <p className="mt-1 text-xs text-slate-600">
                            Maximum bit differences allowed for Bloom filter matches. Lower = stricter. Default: 90.
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
                            placeholder="0.32"
                            className={getInputClass('matching.jaccard_threshold')}
                        />
                        <p className="mt-1 text-xs text-slate-600">
                            Minimum similarity score (0.0-1.0) for MinHash matches. Higher = stricter. Default: 0.5.
                        </p>
                    </div>
                </div>

                {/* Explanation Section */}
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <div className="flex items-start space-x-3">
                        <Info className="h-5 w-5 text-blue-600 mt-0.5" />
                        <div>
                            <h4 className="text-sm font-medium text-blue-900 mb-2">How Matching Works</h4>
                            <div className="space-y-2 text-sm text-blue-800">
                                <p>
                                    <strong>Hamming Threshold:</strong> Measures bit-level differences in Bloom filters.
                                    Lower values are more restrictive (fewer matches but higher precision).
                                </p>
                                <p>
                                    <strong>Jaccard Threshold:</strong> Measures set similarity using MinHash signatures.
                                    Higher values are more restrictive (fewer matches but higher precision).
                                </p>
                                <p className="text-xs italic">
                                    Records must pass BOTH thresholds to be considered a match.
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
} 