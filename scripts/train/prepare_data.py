#!/usr/bin/env python3
"""
Data preparation script for astrology AI training.

This script processes raw astrology data and prepares it for model training.
It handles data cleaning, formatting, and splitting into train/validation/test sets.
"""

import os
import json
import yaml
import pandas as pd
from pathlib import Path
from typing import Dict, List, Any, Optional
from sklearn.model_selection import train_test_split
import logging
from datetime import datetime

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class AstrologyDataPreparer:
    """Handles data preparation for astrology AI training."""

    def __init__(self, config_path: str):
        """Initialize with configuration."""
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)

        self.raw_data_dir = Path(self.config['data']['raw_data_dir'])
        self.processed_data_dir = Path(self.config['data']['processed_data_dir'])
        self.processed_data_dir.mkdir(parents=True, exist_ok=True)

        # Create logs directory
        self.logs_dir = Path("logs")
        self.logs_dir.mkdir(exist_ok=True)

    def load_raw_data(self) -> pd.DataFrame:
        """Load and combine raw astrology data."""
        logger.info("Loading raw astrology data...")

        all_data = []

        # Load Vedic astrology training data
        vedic_file = self.raw_data_dir / "vedic_astrology_training.json"
        if vedic_file.exists():
            try:
                with open(vedic_file, 'r') as f:
                    vedic_data = json.load(f)
                    all_data.extend(vedic_data)
                    logger.info(f"Loaded {len(vedic_data)} Vedic astrology training examples")
            except Exception as e:
                logger.warning(f"Failed to load Vedic data: {e}")

        # Load general astrology data
        general_file = self.raw_data_dir / "astrology_training_data.json"
        if general_file.exists():
            try:
                with open(general_file, 'r') as f:
                    general_data = json.load(f)
                    all_data.extend(general_data)
                    logger.info(f"Loaded {len(general_data)} general astrology training examples")
            except Exception as e:
                logger.warning(f"Failed to load general data: {e}")

        # Load other JSON files from raw directory
        for file_path in self.raw_data_dir.glob("*.json"):
            if file_path.name not in ["vedic_astrology_training.json", "astrology_training_data.json"]:
                try:
                    with open(file_path, 'r') as f:
                        data = json.load(f)
                        if isinstance(data, list):
                            all_data.extend(data)
                        else:
                            all_data.append(data)
                except Exception as e:
                    logger.warning(f"Failed to load {file_path}: {e}")

        # Generate sample data if no data exists
        if not all_data:
            logger.info("No raw data found, generating sample data...")
            all_data = self._generate_sample_data()

        df = pd.DataFrame(all_data)
        logger.info(f"Total loaded {len(df)} records")
        return df

    def _generate_sample_data(self) -> List[Dict[str, Any]]:
        """Generate sample astrology training data."""
        signs = [
            "Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo",
            "Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces"
        ]

        categories = ["daily_horoscope", "weekly_horoscope", "monthly_horoscope", "compatibility", "remedies"]
        sample_data = []

        for sign in signs:
            for category in categories:
                if category == "compatibility":
                    for other_sign in signs:
                        if sign != other_sign:
                            sample_data.append({
                                "input": f"What is the compatibility between {sign} and {other_sign}?",
                                "output": f"{sign} and {other_sign} have {'excellent' if abs(signs.index(sign) - signs.index(other_sign)) % 3 == 0 else 'moderate'} compatibility. The stars suggest...",
                                "category": category,
                                "signs": [sign, other_sign]
                            })
                elif category == "remedies":
                    sample_data.append({
                        "input": f"What remedies should a {sign} follow?",
                        "output": f"For {sign}, focus on these remedies: 1. Wear gemstones like... 2. Practice meditation... 3. Follow dietary guidelines...",
                        "category": category,
                        "signs": [sign]
                    })
                else:
                    sample_data.append({
                        "input": f"Write a {category.replace('_', ' ')} for {sign}.",
                        "output": f"Dear {sign}, the stars indicate that today will bring... Focus on your strengths and...",
                        "category": category,
                        "signs": [sign]
                    })

        return sample_data

    def clean_data(self, df: pd.DataFrame) -> pd.DataFrame:
        """Clean and preprocess the data."""
        logger.info("Cleaning data...")

        # Remove duplicates
        initial_len = len(df)
        df = df.drop_duplicates(subset=['input', 'output'])
        logger.info(f"Removed {initial_len - len(df)} duplicates")

        # Separate Vedic and general data
        vedic_mask = df.apply(lambda row: 'vedic_terms' in row or
                                         (row.get('text', '').startswith(('brihat_', 'phala', 'jataka'))), axis=1)
        vedic_df = df[vedic_mask]
        general_df = df[~vedic_mask]

        logger.info(f"Vedic records: {len(vedic_df)}, General records: {len(general_df)}")

        # Apply different filters
        vedic_config = self.config.get('vedic_astrology', {})
        vedic_filters = vedic_config.get('quality_filters', {})

        # Filter Vedic data (allow empty inputs)
        vedic_min_input = vedic_filters.get('min_input_length', 0)
        vedic_min_output = self.config['data']['filters']['min_output_length']
        vedic_df = vedic_df[vedic_df['input'].str.len() >= vedic_min_input]
        vedic_df = vedic_df[vedic_df['output'].str.len() >= vedic_min_output]

        # Filter general data (use standard filters)
        min_input_len = self.config['data']['filters']['min_input_length']
        min_output_len = self.config['data']['filters']['min_output_length']
        general_df = general_df[general_df['input'].str.len() >= min_input_len]
        general_df = general_df[general_df['output'].str.len() >= min_output_len]

        # Combine back
        df = pd.concat([vedic_df, general_df], ignore_index=True)

        # Apply additional Vedic quality filters
        df = self._apply_vedic_quality_filters(df)

        logger.info(f"Filtered data: {len(df)} records remaining")

        # Add metadata
        df['input_length'] = df['input'].str.len()
        df['output_length'] = df['output'].str.len()
        df['created_at'] = datetime.now().isoformat()

        return df

    def _apply_vedic_quality_filters(self, df: pd.DataFrame) -> pd.DataFrame:
        """Apply Vedic astrology specific quality filters."""
        vedic_config = self.config['vedic_astrology']
        quality_filters = vedic_config.get('quality_filters', {})

        # First, identify Vedic data (data that has vedic_terms or text field indicating Vedic content)
        vedic_mask = df.apply(lambda row: 'vedic_terms' in row or
                                         (row.get('text', '').startswith(('brihat_', 'phala', 'jataka'))), axis=1)

        # Apply Vedic-specific input length filter only to Vedic data
        vedic_min_input = quality_filters.get('min_input_length')
        if vedic_min_input is not None:
            vedic_df = df[vedic_mask]
            non_vedic_df = df[~vedic_mask]

            # Apply different input length requirements
            vedic_df = vedic_df[vedic_df['input'].str.len() >= vedic_min_input]
            non_vedic_df = non_vedic_df[non_vedic_df['input'].str.len() >= self.config['data']['filters']['min_input_length']]

            df = pd.concat([vedic_df, non_vedic_df], ignore_index=True)
            logger.info(f"Applied Vedic input length filter ({vedic_min_input}): {len(df)} records remaining")

        # Filter by Sanskrit terms (only for Vedic data)
        if quality_filters.get('sanskrit_terms_required', 0) > 0:
            sanskrit_terms = quality_filters.get('sanskrit_terms', [])
            min_terms = quality_filters['sanskrit_terms_required']

            def has_sanskrit_terms(row):
                text = str(row.get('output', ''))
                vedic_terms = row.get('vedic_terms', [])
                all_text = text + ' ' + ' '.join(vedic_terms)
                found_terms = sum(1 for term in sanskrit_terms if term.lower() in all_text.lower())
                return found_terms >= min_terms

            vedic_df = df[vedic_mask]
            vedic_df = vedic_df[vedic_df.apply(has_sanskrit_terms, axis=1)]
            df = pd.concat([vedic_df, df[~vedic_mask]], ignore_index=True)
            logger.info(f"Applied Sanskrit terms filter: {len(df)} records remaining")

        # Filter by classical text references (only for Vedic data)
        if quality_filters.get('classical_references_required', False):
            allowed_texts = quality_filters.get('allowed_texts', [])
            vedic_df = df[vedic_mask]
            vedic_df = vedic_df[vedic_df['text'].isin(allowed_texts)]
            df = pd.concat([vedic_df, df[~vedic_mask]], ignore_index=True)
            logger.info(f"Applied classical text filter: {len(df)} records remaining")

        return df

    def format_for_training(self, df: pd.DataFrame) -> List[Dict[str, Any]]:
        """Format data for model training."""
        logger.info("Formatting data for training...")

        formatted_data = []

        for _, row in df.iterrows():
            # Create training example
            example = {
                "instruction": row['input'],
                "input": "",
                "output": row['output'],
                "category": row.get('category', 'general'),
                "signs": row.get('signs', []),
                "metadata": {
                    "input_length": row['input_length'],
                    "output_length": row['output_length'],
                    "created_at": row['created_at']
                }
            }
            formatted_data.append(example)

        return formatted_data

    def split_data(self, data: List[Dict[str, Any]]) -> Dict[str, List[Dict[str, Any]]]:
        """Split data into train/validation/test sets."""
        logger.info("Splitting data into train/val/test sets...")

        # Check if we have enough data for stratified split
        categories = [item.get('category', 'general') for item in data]
        unique_categories = list(set(categories))
        min_samples_per_class = min(categories.count(cat) for cat in unique_categories)

        if len(unique_categories) > 1 and min_samples_per_class >= 2:
            # Stratified split
            try:
                train_data, temp_data = train_test_split(
                    data,
                    test_size=self.config['data']['split']['val_size'] + self.config['data']['split']['test_size'],
                    stratify=categories,
                    random_state=42
                )
                val_data, test_data = train_test_split(
                    temp_data,
                    test_size=self.config['data']['split']['test_size'] / (self.config['data']['split']['val_size'] + self.config['data']['split']['test_size']),
                    stratify=[item.get('category', 'general') for item in temp_data],
                    random_state=42
                )
            except ValueError:
                # Fallback to random split if stratified fails
                logger.warning("Stratified split failed, falling back to random split")
                train_data, temp_data = train_test_split(
                    data,
                    test_size=self.config['data']['split']['val_size'] + self.config['data']['split']['test_size'],
                    random_state=42
                )
                val_data, test_data = train_test_split(
                    temp_data,
                    test_size=self.config['data']['split']['test_size'] / (self.config['data']['split']['val_size'] + self.config['data']['split']['test_size']),
                    random_state=42
                )
        else:
            # Random split
            train_data, temp_data = train_test_split(
                data,
                test_size=self.config['data']['split']['val_size'] + self.config['data']['split']['test_size'],
                random_state=42
            )
            val_data, test_data = train_test_split(
                temp_data,
                test_size=self.config['data']['split']['test_size'] / (self.config['data']['split']['val_size'] + self.config['data']['split']['test_size']),
                random_state=42
            )

        logger.info(f"Split sizes: train={len(train_data)}, val={len(val_data)}, test={len(test_data)}")

        return {
            'train': train_data,
            'validation': val_data,
            'test': test_data
        }

    def save_data(self, split_data: Dict[str, List[Dict[str, Any]]]):
        """Save processed data to files."""
        logger.info("Saving processed data...")

        for split_name, data in split_data.items():
            output_file = self.processed_data_dir / f"{split_name}.jsonl"
            with open(output_file, 'w') as f:
                for item in data:
                    f.write(json.dumps(item, ensure_ascii=False) + '\n')

            logger.info(f"Saved {len(data)} records to {output_file}")

        # Save statistics
        stats = {
            'total_records': sum(len(data) for data in split_data.values()),
            'splits': {name: len(data) for name, data in split_data.items()},
            'categories': {},
            'created_at': datetime.now().isoformat()
        }

        # Calculate category distribution
        for split_name, data in split_data.items():
            category_counts = {}
            for item in data:
                cat = item.get('category', 'general')
                category_counts[cat] = category_counts.get(cat, 0) + 1
            stats['categories'][split_name] = category_counts

        stats_file = self.processed_data_dir / "data_stats.json"
        with open(stats_file, 'w') as f:
            json.dump(stats, f, indent=2)

        logger.info(f"Saved statistics to {stats_file}")

    def run(self):
        """Run the complete data preparation pipeline."""
        logger.info("Starting data preparation pipeline...")

        try:
            # Load raw data
            df = self.load_raw_data()

            # Clean data
            df = self.clean_data(df)

            # Format for training
            formatted_data = self.format_for_training(df)

            # Split data
            split_data = self.split_data(formatted_data)

            # Save processed data
            self.save_data(split_data)

            logger.info("Data preparation completed successfully!")

        except Exception as e:
            logger.error(f"Data preparation failed: {e}")
            raise

def main():
    """Main entry point."""
    import argparse

    parser = argparse.ArgumentParser(description="Prepare astrology data for training")
    parser.add_argument("--config", default="config/data_config.yaml", help="Path to config file")
    args = parser.parse_args()

    preparer = AstrologyDataPreparer(args.config)
    preparer.run()

if __name__ == "__main__":
    main()