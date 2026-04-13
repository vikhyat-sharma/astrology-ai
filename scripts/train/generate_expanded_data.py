#!/usr/bin/env python3
"""
Expanded astrology training data generator.

This script generates comprehensive training data for the astrology AI model,
covering various astrology topics, signs, and concepts.
"""

import json
import random
from datetime import datetime
from typing import List, Dict, Any
from pathlib import Path

class AstrologyDataGenerator:
    """Generates comprehensive astrology training data."""

    def __init__(self):
        self.zodiac_signs = [
            "Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo",
            "Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces"
        ]

        self.planets = [
            "Sun", "Moon", "Mars", "Mercury", "Jupiter", "Venus", "Saturn", "Rahu", "Ketu"
        ]

        self.categories = [
            "daily_horoscope", "weekly_horoscope", "monthly_horoscope",
            "compatibility", "remedies", "planetary_analysis", "house_analysis",
            "yoga_identification", "birth_chart_reading", "dasha_prediction",
            "predictive_methods", "vedic_astrology"
        ]

    def generate_daily_horoscopes(self) -> List[Dict[str, Any]]:
        """Generate daily horoscopes for all signs."""
        horoscopes = []

        daily_themes = [
            "career advancement and leadership",
            "love and relationships",
            "financial opportunities",
            "health and wellness",
            "spiritual growth",
            "creative expression",
            "social connections",
            "personal transformation",
            "learning and communication",
            "family matters",
            "travel and adventure",
            "rest and rejuvenation"
        ]

        for sign in self.zodiac_signs:
            for theme in random.sample(daily_themes, 3):  # Generate 3 per sign
                instruction = f"Write a daily horoscope for {sign}."
                output = f"Dear {sign}, today brings opportunities for {theme}. The stars align to support your natural tendencies and help you achieve your goals. Trust your instincts and take positive action. Financial matters show promise with careful planning. Love brings warmth and connection to your relationships. Health remains strong with balanced lifestyle choices. Remember your unique strengths and use them to navigate any challenges that arise."

                horoscopes.append({
                    "instruction": instruction,
                    "input": "",
                    "output": output,
                    "category": "daily_horoscope",
                    "signs": [sign],
                    "metadata": {
                        "input_length": len(instruction),
                        "output_length": len(output),
                        "created_at": datetime.now().isoformat()
                    }
                })

        return horoscopes

    def generate_weekly_horoscopes(self) -> List[Dict[str, Any]]:
        """Generate weekly horoscopes for all signs."""
        horoscopes = []

        for sign in self.zodiac_signs:
            instruction = f"Write a weekly horoscope for {sign}."
            output = f"Dear {sign}, this week emphasizes growth and new beginnings. Monday-Thursday: Focus on career and professional development. Friday-Sunday: Personal relationships and creative pursuits take center stage. Your ruling planet's influence brings clarity and purpose. Financial opportunities arise through smart planning. Love relationships deepen with open communication. Health benefits from regular exercise and balanced nutrition. Trust your intuition and embrace opportunities for positive change."

            horoscopes.append({
                "instruction": instruction,
                "input": "",
                "output": output,
                "category": "weekly_horoscope",
                "signs": [sign],
                "metadata": {
                    "input_length": len(instruction),
                    "output_length": len(output),
                    "created_at": datetime.now().isoformat()
                }
            })

        return horoscopes

    def generate_monthly_horoscopes(self) -> List[Dict[str, Any]]:
        """Generate monthly horoscopes for all signs."""
        horoscopes = []

        for sign in self.zodiac_signs:
            instruction = f"Write a monthly horoscope for {sign}."
            output = f"Dear {sign}, {datetime.now().strftime('%B')} brings significant opportunities for growth and achievement. The planetary alignments support your goals and aspirations. Career advancement comes through dedication and hard work. Relationships deepen with mutual understanding and compromise. Financial stability improves through careful planning and wise investments. Health and wellness benefit from holistic approaches. Creative projects flourish with your natural talents. Trust the universe's timing and embrace positive changes."

            horoscopes.append({
                "instruction": instruction,
                "input": "",
                "output": output,
                "category": "monthly_horoscope",
                "signs": [sign],
                "metadata": {
                    "input_length": len(instruction),
                    "output_length": len(output),
                    "created_at": datetime.now().isoformat()
                }
            })

        return horoscopes

    def generate_compatibility_data(self) -> List[Dict[str, Any]]:
        """Generate compatibility analysis for sign combinations."""
        compatibility_data = []

        compatibility_matrix = {
            ("Aries", "Leo"): "excellent fire sign compatibility",
            ("Aries", "Sagittarius"): "adventurous and dynamic",
            ("Taurus", "Virgo"): "practical and stable",
            ("Taurus", "Capricorn"): "reliable and ambitious",
            ("Gemini", "Libra"): "intellectual and communicative",
            ("Gemini", "Aquarius"): "innovative and progressive",
            ("Cancer", "Scorpio"): "emotional and intuitive",
            ("Cancer", "Pisces"): "nurturing and compassionate",
            ("Leo", "Sagittarius"): "dramatic and enthusiastic",
            ("Virgo", "Capricorn"): "disciplined and organized",
            ("Libra", "Aquarius"): "harmonious and idealistic",
            ("Scorpio", "Pisces"): "mysterious and spiritual"
        }

        for (sign1, sign2), description in compatibility_matrix.items():
            instruction = f"What is the compatibility between {sign1} and {sign2}?"
            output = f"{sign1} and {sign2} share {description}, creating a strong foundation for lasting relationships. Both signs bring complementary qualities that enhance their connection. {sign1} contributes enthusiasm and initiative, while {sign2} adds depth and stability. Together, they can achieve great things through mutual support and understanding. Communication flows naturally, and shared values create lasting bonds. In love, this pairing is passionate and fulfilling. Career-wise, they make excellent partners in creative or leadership roles."

            compatibility_data.append({
                "instruction": instruction,
                "input": "",
                "output": output,
                "category": "compatibility",
                "signs": [sign1, sign2],
                "metadata": {
                    "input_length": len(instruction),
                    "output_length": len(output),
                    "created_at": datetime.now().isoformat()
                }
            })

        return compatibility_data

    def generate_remedies_data(self) -> List[Dict[str, Any]]:
        """Generate remedial measures for all signs."""
        remedies_data = []

        remedies_by_sign = {
            "Aries": ["Wear red coral for courage", "Practice warrior yoga poses", "Use cinnamon essential oil", "Chant 'RAM' mantra", "Practice grounding exercises"],
            "Taurus": ["Wear emerald for stability", "Practice earth-based yoga", "Use rose essential oil", "Chant 'SHAM' mantra", "Create stable routines"],
            "Gemini": ["Wear emerald for communication", "Practice pranayama breathing", "Use lavender essential oil", "Chant 'YAM' mantra", "Keep a journal"],
            "Cancer": ["Wear pearl for emotional balance", "Practice moon salutations", "Use jasmine essential oil", "Chant 'SHAM' mantra", "Create emotional boundaries"],
            "Leo": ["Wear ruby for confidence", "Practice sun salutations", "Use orange essential oil", "Chant 'RAM' mantra", "Practice self-expression"],
            "Virgo": ["Wear emerald for health", "Practice grounding poses", "Use peppermint essential oil", "Chant 'SHAM' mantra", "Practice self-care rituals"],
            "Libra": ["Wear diamond for harmony", "Practice balance poses", "Use rose essential oil", "Chant 'SHAM' mantra", "Practice meditation"],
            "Scorpio": ["Wear coral for transformation", "Practice intense yoga", "Use patchouli essential oil", "Chant 'VAM' mantra", "Practice shadow work"],
            "Sagittarius": ["Wear yellow sapphire", "Practice fire-based activities", "Use frankincense oil", "Chant 'OM' mantra", "Travel and explore"],
            "Capricorn": ["Wear blue sapphire", "Practice mountain pose", "Use cedarwood oil", "Chant 'VAM' mantra", "Set clear goals"],
            "Aquarius": ["Wear amethyst for innovation", "Practice air-based yoga", "Use eucalyptus oil", "Chant 'YAM' mantra", "Practice community service"],
            "Pisces": ["Wear moonstone for intuition", "Practice water-based yoga", "Use sandalwood oil", "Chant 'SHAM' mantra", "Practice creative visualization"]
        }

        for sign, remedies in remedies_by_sign.items():
            instruction = f"What astrological remedies should a {sign} follow?"
            remedies_text = "For " + sign + ", embrace these remedies: " + ", ".join(f"{i+1}. {remedy}" for i, remedy in enumerate(remedies))
            output = remedies_text + f". Remember, your unique {sign.lower()} qualities are your greatest strength when balanced properly."

            remedies_data.append({
                "instruction": instruction,
                "input": "",
                "output": output,
                "category": "remedies",
                "signs": [sign],
                "metadata": {
                    "input_length": len(instruction),
                    "output_length": len(output),
                    "created_at": datetime.now().isoformat()
                }
            })

        return remedies_data

    def generate_planetary_analysis(self) -> List[Dict[str, Any]]:
        """Generate planetary analysis data."""
        planetary_data = []

        planet_info = {
            "Sun": "represents soul, father, power, leadership, government, self, vitality",
            "Moon": "represents mind, mother, emotions, public, receptivity, intuition",
            "Mars": "represents energy, siblings, courage, land, passion, action",
            "Mercury": "represents intelligence, communication, business, logic, learning",
            "Jupiter": "represents wisdom, spirituality, wealth, teaching, expansion",
            "Venus": "represents love, beauty, arts, luxury, harmony, relationships",
            "Saturn": "represents discipline, karma, hard work, structure, limitations",
            "Rahu": "represents ambition, foreign lands, unconventional paths, innovation",
            "Ketu": "represents spirituality, detachment, past life karma, liberation"
        }

        for planet, description in planet_info.items():
            instruction = f"Explain the significance of {planet} in Vedic astrology."
            output = f"In Vedic astrology, {planet} {description}. According to classical texts like Brihat Parashara Hora Shastra, {planet} governs specific areas of life and has unique characteristics. When well-placed, {planet} brings positive results in its domains. When challenged, it indicates areas requiring attention and growth. The planet's strength depends on its sign placement, aspects from other planets, and current dasha periods."

            planetary_data.append({
                "instruction": instruction,
                "input": "",
                "output": output,
                "category": "planetary_analysis",
                "signs": [],
                "metadata": {
                    "input_length": len(instruction),
                    "output_length": len(output),
                    "created_at": datetime.now().isoformat()
                }
            })

        return planetary_data

    def generate_house_analysis(self) -> List[Dict[str, Any]]:
        """Generate house analysis data."""
        house_data = []

        houses_info = {
            1: "Self, personality, physical appearance, first impressions, overall life direction",
            2: "Wealth, family, speech, food, material possessions, self-worth",
            3: "Siblings, communication, short journeys, courage, skills, learning",
            4: "Home, mother, property, education, emotional foundation, happiness",
            5: "Children, creativity, intelligence, spiritual practices, romance",
            6: "Health, enemies, service, daily routine, obstacles, purification",
            7: "Marriage, partnerships, spouse, business relationships, harmony",
            8: "Longevity, secrets, occult, transformation, inheritance, intimacy",
            9: "Father, guru, dharma, long journeys, fortune, higher learning",
            10: "Career, reputation, authority, public life, achievements, status",
            11: "Gains, elder siblings, hopes, wishes, friends, fulfillment",
            12: "Expenses, foreign lands, spirituality, losses, isolation, enlightenment"
        }

        for house_num, significance in houses_info.items():
            instruction = f"What does the {house_num}{('st' if house_num == 1 else 'nd' if house_num == 2 else 'rd' if house_num == 3 else 'th')} house represent in astrology?"
            output = f"The {house_num}{('st' if house_num == 1 else 'nd' if house_num == 2 else 'rd' if house_num == 3 else 'th')} house represents {significance}. In Vedic astrology, houses are areas of life experience and karma. Planets placed in this house influence these life areas. The house lord's strength and planetary aspects modify the results. A strong {house_num}{('st' if house_num == 1 else 'nd' if house_num == 2 else 'rd' if house_num == 3 else 'th')} house brings positive outcomes in its domains, while challenges indicate areas for spiritual growth."

            house_data.append({
                "instruction": instruction,
                "input": "",
                "output": output,
                "category": "house_analysis",
                "signs": [],
                "metadata": {
                    "input_length": len(instruction),
                    "output_length": len(output),
                    "created_at": datetime.now().isoformat()
                }
            })

        return house_data

    def generate_yoga_analysis(self) -> List[Dict[str, Any]]:
        """Generate yoga analysis data."""
        yoga_data = []

        yogas = [
            {
                "name": "Raja Yoga",
                "description": "planetary combinations for power, authority, and leadership",
                "combinations": ["lords of Kendra and Trikona houses combine", "exalted planets in angular houses", "benefics strong in 10th house"]
            },
            {
                "name": "Dhana Yoga",
                "description": "wealth and prosperity combinations",
                "combinations": ["2nd lord strong with benefics", "11th lord with 9th or 10th lord", "Jupiter aspects 2nd house"]
            },
            {
                "name": "Gaja Kesari Yoga",
                "description": "wisdom and prosperity combination",
                "combinations": ["Moon and Jupiter in mutual aspect", "Moon and Jupiter in kendra from each other"]
            },
            {
                "name": "Panchmahapurusha Yoga",
                "description": "great personality combinations",
                "combinations": ["five specific planetary placements for exceptional qualities"]
            }
        ]

        for yoga in yogas:
            instruction = f"Explain {yoga['name']} in Vedic astrology."
            combinations_text = ", ".join(yoga['combinations'])
            output = f"{yoga['name']} is a powerful {yoga['description']}. Key combinations include: {combinations_text}. According to classical texts, this yoga brings exceptional results in its domain. The strength depends on planetary dignity, aspects, and dasha periods. When active, it significantly enhances life outcomes and spiritual growth."

            yoga_data.append({
                "instruction": instruction,
                "input": "",
                "output": output,
                "category": "yoga_identification",
                "signs": [],
                "metadata": {
                    "input_length": len(instruction),
                    "output_length": len(output),
                    "created_at": datetime.now().isoformat()
                }
            })

        return yoga_data

    def generate_birth_chart_readings(self) -> List[Dict[str, Any]]:
        """Generate birth chart interpretation examples."""
        chart_data = []

        chart_examples = [
            "Sun in Leo with Moon in Cancer",
            "Jupiter in Sagittarius with Venus in Pisces",
            "Mars in Aries with Saturn in Capricorn",
            "Mercury in Gemini with Moon in Virgo"
        ]

        for chart_combo in chart_examples:
            instruction = f"Interpret this birth chart combination: {chart_combo}."
            output = f"This combination creates a powerful and dynamic chart. {chart_combo} suggests strong leadership qualities combined with emotional intelligence. The planetary placements indicate natural talent for creative expression and spiritual growth. According to Vedic principles, this creates favorable yogas for success and fulfillment. The individual would excel in fields requiring both intuition and practical application."

            chart_data.append({
                "instruction": instruction,
                "input": "",
                "output": output,
                "category": "birth_chart_reading",
                "signs": [],
                "metadata": {
                    "input_length": len(instruction),
                    "output_length": len(output),
                    "created_at": datetime.now().isoformat()
                }
            })

        return chart_data

    def generate_dasha_predictions(self) -> List[Dict[str, Any]]:
        """Generate dasha prediction examples."""
        dasha_data = []

        dasha_periods = [
            "Sun dasha (6 years)",
            "Moon dasha (10 years)",
            "Mars dasha (7 years)",
            "Rahu dasha (18 years)",
            "Jupiter dasha (16 years)",
            "Saturn dasha (19 years)",
            "Mercury dasha (17 years)",
            "Ketu dasha (7 years)",
            "Venus dasha (20 years)"
        ]

        for dasha in dasha_periods:
            planet = dasha.split()[0]
            instruction = f"What can be expected during {dasha}?"
            output = f"During {dasha}, the energies of {planet} become prominent in life. {planet} governs specific areas and brings corresponding experiences. The results depend on {planet}'s placement, aspects, and dignity in the birth chart. Favorable placements bring positive outcomes in {planet}'s domains, while challenging placements may bring lessons and growth opportunities. This period emphasizes {planet}'s qualities and life areas."

            dasha_data.append({
                "instruction": instruction,
                "input": "",
                "output": output,
                "category": "dasha_prediction",
                "signs": [],
                "metadata": {
                    "input_length": len(instruction),
                    "output_length": len(output),
                    "created_at": datetime.now().isoformat()
                }
            })

        return dasha_data

    def generate_all_data(self) -> List[Dict[str, Any]]:
        """Generate all training data."""
        all_data = []

        # Generate data from all methods
        generators = [
            self.generate_daily_horoscopes,
            self.generate_weekly_horoscopes,
            self.generate_monthly_horoscopes,
            self.generate_compatibility_data,
            self.generate_remedies_data,
            self.generate_planetary_analysis,
            self.generate_house_analysis,
            self.generate_yoga_analysis,
            self.generate_birth_chart_readings,
            self.generate_dasha_predictions
        ]

        for generator in generators:
            all_data.extend(generator())

        return all_data

    def save_data(self, data: List[Dict[str, Any]], output_file: str):
        """Save generated data to JSONL file."""
        with open(output_file, 'w', encoding='utf-8') as f:
            for item in data:
                json.dump(item, f, ensure_ascii=False)
                f.write('\n')

        print(f"Generated {len(data)} training examples and saved to {output_file}")

def main():
    """Main function to generate expanded training data."""
    generator = AstrologyDataGenerator()
    data = generator.generate_all_data()

    # Save to raw data directory
    output_dir = Path("data/raw")
    output_dir.mkdir(parents=True, exist_ok=True)

    output_file = output_dir / "expanded_astrology_training.jsonl"
    generator.save_data(data, str(output_file))

    print(f"Total training examples: {len(data)}")

if __name__ == "__main__":
    main()